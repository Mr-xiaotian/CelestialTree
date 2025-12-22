#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
CelestialTree 压测脚本（多场景）
- emit：写入事件（可控 payload 大小、parents 数量、分叉/合流）
- get/children/heads/descendants：读取接口压测
- mixed：读写混合
- sse：订阅连接数 + 写入触发推送（粗略观测 fan-out 成本）

依赖：
  pip install httpx

用法示例：
  python bench_celestialtree.py --base http://127.0.0.1:7777 scenario emit --n 20000 --c 200 --payload-bytes 256
  python bench_celestialtree.py --base http://127.0.0.1:7777 scenario mixed --seconds 15 --c 200 --write-ratio 0.2
  python bench_celestialtree.py --base http://127.0.0.1:7777 scenario descendants --seconds 10 --c 50
  python bench_celestialtree.py --base http://127.0.0.1:7777 scenario sse --subs 300 --emit-rate 200 --seconds 10
"""

from __future__ import annotations

import argparse
import asyncio
import json
import math
import os
import random
import statistics
import string
import time
from dataclasses import dataclass
from typing import Any, Dict, List, Optional, Tuple

import httpx


# ---------------------------
# Helpers
# ---------------------------

def now_ms() -> float:
    return time.perf_counter() * 1000.0


def human(n: float) -> str:
    if n >= 1e9:
        return f"{n/1e9:.2f}G"
    if n >= 1e6:
        return f"{n/1e6:.2f}M"
    if n >= 1e3:
        return f"{n/1e3:.2f}K"
    return f"{n:.0f}"


def pct(values: List[float], p: float) -> float:
    """percentile (0-100) with linear interpolation"""
    if not values:
        return float("nan")
    v = sorted(values)
    if len(v) == 1:
        return v[0]
    k = (len(v) - 1) * (p / 100.0)
    f = math.floor(k)
    c = math.ceil(k)
    if f == c:
        return v[int(k)]
    return v[f] + (v[c] - v[f]) * (k - f)


def make_payload(payload_bytes: int) -> Dict[str, Any]:
    # 生成指定大小左右的字符串；payload 结构固定，便于对比
    # 注意：json 编码后会有额外开销，所以这里只做近似
    if payload_bytes <= 0:
        return {"msg": "hi"}
    s = "".join(random.choice(string.ascii_letters + string.digits) for _ in range(payload_bytes))
    return {"blob": s}


@dataclass
class Stats:
    ok: int = 0
    fail: int = 0
    lat_ms: List[float] = None

    def __post_init__(self):
        if self.lat_ms is None:
            self.lat_ms = []

    def add_ok(self, lat: float):
        self.ok += 1
        self.lat_ms.append(lat)

    def add_fail(self, lat: float):
        self.fail += 1
        self.lat_ms.append(lat)

    def summary(self, elapsed_s: float) -> str:
        total = self.ok + self.fail
        rps = total / elapsed_s if elapsed_s > 0 else 0.0
        ok_rps = self.ok / elapsed_s if elapsed_s > 0 else 0.0
        lat = self.lat_ms
        avg = statistics.mean(lat) if lat else float("nan")
        p50 = pct(lat, 50)
        p90 = pct(lat, 90)
        p95 = pct(lat, 95)
        p99 = pct(lat, 99)
        mx = max(lat) if lat else float("nan")
        return (
            f"total={human(total)} ok={human(self.ok)} fail={human(self.fail)} "
            f"rps={rps:.1f} ok_rps={ok_rps:.1f} "
            f"lat_ms(avg={avg:.2f} p50={p50:.2f} p90={p90:.2f} p95={p95:.2f} p99={p99:.2f} max={mx:.2f})"
        )


# ---------------------------
# CelestialTree Client
# ---------------------------

class CTClient:
    def __init__(self, base: str, timeout: float = 10.0):
        self.base = base.rstrip("/")
        limits = httpx.Limits(max_connections=2000, max_keepalive_connections=2000)
        self.client = httpx.AsyncClient(timeout=timeout, limits=limits, http2=False)

    async def close(self):
        await self.client.aclose()

    async def healthz(self) -> Dict[str, Any]:
        r = await self.client.get(f"{self.base}/healthz")
        r.raise_for_status()
        return r.json()

    async def heads(self) -> List[int]:
        r = await self.client.get(f"{self.base}/heads")
        r.raise_for_status()
        return r.json().get("heads", [])

    async def get_event(self, eid: int) -> Dict[str, Any]:
        r = await self.client.get(f"{self.base}/event/{eid}")
        r.raise_for_status()
        return r.json()

    async def children(self, eid: int) -> List[int]:
        r = await self.client.get(f"{self.base}/children/{eid}")
        r.raise_for_status()
        return r.json().get("children", [])

    async def descendants(self, eid: int) -> Dict[str, Any]:
        r = await self.client.get(f"{self.base}/descendants/{eid}")
        r.raise_for_status()
        return r.json()

    async def emit(self, typ: str, parents: List[int], payload: Dict[str, Any], meta: Dict[str, Any]) -> Dict[str, Any]:
        r = await self.client.post(
            f"{self.base}/emit",
            json={"type": typ, "parents": parents, "payload": payload, "meta": meta},
        )
        r.raise_for_status()
        return r.json()


# ---------------------------
# Workload builders
# ---------------------------

async def ensure_genesis_heads(ct: CTClient) -> List[int]:
    # 尝试 heads；若空就靠服务启动时的 genesis（你的 main.go 已经做了）
    heads = await ct.heads()
    return heads


def pick_parents(mode: str, pool: List[int], k: int) -> List[int]:
    """
    mode:
      - chain:   只选一个 parent（构成链）
      - fork:    多选（从 pool 里抽）制造分叉/并行分支
      - merge:   k>=2，从不同分支抽，制造合流
      - random:  随机
    """
    if not pool:
        return []
    if k <= 0:
        return []
    if mode == "chain":
        return [pool[-1]]
    if mode == "fork":
        # 从最近的头部附近抽，制造“向前生长”的分叉
        n = min(len(pool), max(5, k * 10))
        base = pool[-n:]
        return random.sample(base, k=min(k, len(base)))
    if mode == "merge":
        n = min(len(pool), max(20, k * 20))
        base = pool[-n:]
        return random.sample(base, k=min(k, len(base)))
    # random
    return random.sample(pool, k=min(k, len(pool)))


# ---------------------------
# Scenarios
# ---------------------------

async def scenario_emit(args) -> None:
    ct = CTClient(args.base, timeout=args.timeout)
    try:
        await ct.healthz()
        heads = await ensure_genesis_heads(ct)
        # 维护一个“可作为 parent 的池”
        pool: List[int] = heads[:]

        sem = asyncio.Semaphore(args.c)
        stats = Stats()
        start = time.perf_counter()

        async def one(i: int):
            nonlocal pool
            async with sem:
                t0 = now_ms()
                try:
                    parents = pick_parents(args.parents_mode, pool, args.parents_k)
                    payload = make_payload(args.payload_bytes)
                    meta = {"bench": "emit", "i": i}
                    out = await ct.emit(args.type, parents, payload, meta)
                    eid = int(out["event"]["id"])
                    # 更新 pool（不加锁也行：近似即可；想严谨可加 asyncio.Lock）
                    pool.append(eid)
                    stats.add_ok(now_ms() - t0)
                except Exception:
                    stats.add_fail(now_ms() - t0)

        tasks = [asyncio.create_task(one(i)) for i in range(args.n)]
        await asyncio.gather(*tasks)

        elapsed = time.perf_counter() - start
        print("[emit]", stats.summary(elapsed))
        print(f"[emit] pool_size={len(pool)} last_id={pool[-1] if pool else 'n/a'}")
    finally:
        await ct.close()


async def scenario_read(args) -> None:
    ct = CTClient(args.base, timeout=args.timeout)
    try:
        await ct.healthz()
        heads = await ensure_genesis_heads(ct)
        if not heads:
            raise RuntimeError("heads is empty; emit something first.")
        # 做一点预热：拿到一些可读 ID
        pool = heads[:]
        # 扩充池子：从 heads 向下取 children
        for _ in range(min(5, args.warm_depth)):
            new_ids = []
            for eid in pool[-min(len(pool), 20):]:
                try:
                    ch = await ct.children(eid)
                    new_ids.extend(ch)
                except Exception:
                    pass
            if new_ids:
                pool.extend(new_ids)

        sem = asyncio.Semaphore(args.c)
        stats = Stats()
        start = time.perf_counter()

        async def one():
            async with sem:
                eid = random.choice(pool)
                t0 = now_ms()
                try:
                    if args.read_op == "event":
                        await ct.get_event(eid)
                    elif args.read_op == "children":
                        await ct.children(eid)
                    elif args.read_op == "heads":
                        await ct.heads()
                    elif args.read_op == "descendants":
                        await ct.descendants(eid)
                    else:
                        raise ValueError("bad read_op")
                    stats.add_ok(now_ms() - t0)
                except Exception:
                    stats.add_fail(now_ms() - t0)

        if args.seconds > 0:
            end_at = time.perf_counter() + args.seconds
            tasks = []
            while time.perf_counter() < end_at:
                tasks.append(asyncio.create_task(one()))
                # 允许排队，但别无穷堆积
                if len(tasks) >= args.c * 4:
                    await asyncio.gather(*tasks)
                    tasks.clear()
            if tasks:
                await asyncio.gather(*tasks)
        else:
            tasks = [asyncio.create_task(one()) for _ in range(args.n)]
            await asyncio.gather(*tasks)

        elapsed = time.perf_counter() - start
        print(f"[read:{args.read_op}]", stats.summary(elapsed))
    finally:
        await ct.close()


async def scenario_mixed(args) -> None:
    ct = CTClient(args.base, timeout=args.timeout)
    try:
        await ct.healthz()
        heads = await ensure_genesis_heads(ct)
        pool: List[int] = heads[:] if heads else []

        sem = asyncio.Semaphore(args.c)
        stats_w = Stats()
        stats_r = Stats()
        start = time.perf_counter()
        end_at = start + args.seconds

        # 简单策略：pool 只追加（不严格同步，够压测）
        async def do_write(i: int):
            nonlocal pool
            async with sem:
                t0 = now_ms()
                try:
                    parents = pick_parents(args.parents_mode, pool, args.parents_k)
                    payload = make_payload(args.payload_bytes)
                    meta = {"bench": "mixed", "kind": "write", "i": i}
                    out = await ct.emit(args.type, parents, payload, meta)
                    eid = int(out["event"]["id"])
                    pool.append(eid)
                    stats_w.add_ok(now_ms() - t0)
                except Exception:
                    stats_w.add_fail(now_ms() - t0)

        async def do_read(i: int):
            async with sem:
                t0 = now_ms()
                try:
                    if not pool:
                        await ct.heads()
                    else:
                        eid = random.choice(pool)
                        op = random.choice(args.read_ops.split(","))
                        if op == "event":
                            await ct.get_event(eid)
                        elif op == "children":
                            await ct.children(eid)
                        elif op == "heads":
                            await ct.heads()
                        elif op == "descendants":
                            await ct.descendants(eid)
                        else:
                            await ct.get_event(eid)
                    stats_r.add_ok(now_ms() - t0)
                except Exception:
                    stats_r.add_fail(now_ms() - t0)

        i = 0
        tasks: List[asyncio.Task] = []
        while time.perf_counter() < end_at:
            i += 1
            if random.random() < args.write_ratio:
                tasks.append(asyncio.create_task(do_write(i)))
            else:
                tasks.append(asyncio.create_task(do_read(i)))
            if len(tasks) >= args.c * 6:
                await asyncio.gather(*tasks)
                tasks.clear()

        if tasks:
            await asyncio.gather(*tasks)

        elapsed = time.perf_counter() - start
        print("[mixed] write:", stats_w.summary(elapsed))
        print("[mixed] read :", stats_r.summary(elapsed))
        print(f"[mixed] pool_size={len(pool)}")
    finally:
        await ct.close()


async def scenario_descendants(args) -> None:
    # 先构建一个相对“可观的”子树，再打 descendants
    ct = CTClient(args.base, timeout=args.timeout)
    try:
        await ct.healthz()
        heads = await ensure_genesis_heads(ct)
        pool = heads[:]
        if not pool:
            raise RuntimeError("heads empty; service not initialized?")

        # 构树：通过 fork/merge 生成一些结构
        build_n = args.build_n
        build_c = min(args.c, args.build_c)
        sem = asyncio.Semaphore(build_c)

        async def build_one(i: int):
            nonlocal pool
            async with sem:
                try:
                    parents = pick_parents(args.parents_mode, pool, args.parents_k)
                    payload = make_payload(args.payload_bytes)
                    meta = {"bench": "descendants-build", "i": i}
                    out = await ct.emit("node", parents, payload, meta)
                    pool.append(int(out["event"]["id"]))
                except Exception:
                    pass

        await asyncio.gather(*[asyncio.create_task(build_one(i)) for i in range(build_n)])

        # 选一个较早的节点作为根，更可能有后代
        root = pool[max(0, len(pool)//3)]

        # 压 descendants
        sem2 = asyncio.Semaphore(args.c)
        stats = Stats()
        start = time.perf_counter()
        end_at = start + args.seconds

        async def one():
            async with sem2:
                t0 = now_ms()
                try:
                    await ct.descendants(root)
                    stats.add_ok(now_ms() - t0)
                except Exception:
                    stats.add_fail(now_ms() - t0)

        tasks: List[asyncio.Task] = []
        while time.perf_counter() < end_at:
            tasks.append(asyncio.create_task(one()))
            if len(tasks) >= args.c * 4:
                await asyncio.gather(*tasks)
                tasks.clear()
        if tasks:
            await asyncio.gather(*tasks)

        elapsed = time.perf_counter() - start
        print(f"[descendants] root={root} build_n={build_n} pool={len(pool)}")
        print("[descendants]", stats.summary(elapsed))
    finally:
        await ct.close()


async def scenario_sse(args) -> None:
    """
    SSE 粗压测：
    - 先建立 N 个订阅连接
    - 再以 emit_rate 触发写入
    - 统计每个订阅者实际收到的 emit 事件数（由于你服务端会丢慢订阅者事件，这是预期行为）
    """
    base = args.base.rstrip("/")
    subscribe_url = f"{base}/subscribe"

    async def one_sub(idx: int) -> int:
        received = 0
        async with httpx.AsyncClient(timeout=None) as client:
            async with client.stream("GET", subscribe_url, headers={"Accept": "text/event-stream"}) as r:
                # 简单解析 SSE：按空行分隔事件块
                buf_lines: List[str] = []
                async for line in r.aiter_lines():
                    if line is None:
                        continue
                    if line == "":
                        # end of event
                        # event: emit
                        # data: {...}
                        evt = "\n".join(buf_lines)
                        buf_lines.clear()
                        if "event: emit" in evt:
                            received += 1
                        # 退出条件：由外层控制（超时取消任务）
                        continue
                    buf_lines.append(line)
        return received

    # 先建立订阅者（并发启动）
    print(f"[sse] connecting subs={args.subs} ...")
    sub_tasks = [asyncio.create_task(one_sub(i)) for i in range(args.subs)]
    await asyncio.sleep(0.5)
    print("[sse] subs connected (or in-progress). start emitting...")

    # 发射写入
    ct = CTClient(args.base, timeout=args.timeout)
    try:
        await ct.healthz()
        heads = await ensure_genesis_heads(ct)
        pool: List[int] = heads[:]

        start = time.perf_counter()
        end_at = start + args.seconds
        emitted = 0
        interval = 1.0 / max(1, args.emit_rate)

        while time.perf_counter() < end_at:
            t0 = time.perf_counter()
            # 简单写入：chain 模式保证 parent 存在且增长
            try:
                parents = [pool[-1]] if pool else []
                out = await ct.emit("sse", parents, {"msg": "tick"}, {"bench": "sse"})
                pool.append(int(out["event"]["id"]))
                emitted += 1
            except Exception:
                pass

            dt = time.perf_counter() - t0
            sleep_for = max(0.0, interval - dt)
            if sleep_for > 0:
                await asyncio.sleep(sleep_for)

        # 结束：取消订阅任务
        for t in sub_tasks:
            t.cancel()
        results = []
        for t in sub_tasks:
            try:
                results.append(await t)
            except asyncio.CancelledError:
                # 取消时没返回，算 0
                results.append(0)
            except Exception:
                results.append(0)

        # 统计
        total_recv = sum(results)
        avg_recv = total_recv / len(results) if results else 0.0
        min_recv = min(results) if results else 0
        max_recv = max(results) if results else 0
        p50 = pct([float(x) for x in results], 50)
        p90 = pct([float(x) for x in results], 90)

        print(f"[sse] emitted={emitted} subs={args.subs}")
        print(f"[sse] recv_total={total_recv} recv_avg={avg_recv:.1f} recv_min={min_recv} recv_max={max_recv} p50={p50:.0f} p90={p90:.0f}")
        print("[sse] 注意：你的服务端设计允许慢订阅者丢事件，这里收到数低是正常现象。")
    finally:
        await ct.close()


# ---------------------------
# CLI
# ---------------------------

def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="CelestialTree benchmark tool")
    p.add_argument("--base", default="http://127.0.0.1:7777", help="server base url, e.g. http://127.0.0.1:7777")
    p.add_argument("--timeout", type=float, default=10.0, help="http timeout seconds")
    sub = p.add_subparsers(dest="cmd", required=True)

    # scenario
    sp = sub.add_parser("scenario", help="run a benchmark scenario")
    sp.add_argument("name", choices=["emit", "read", "mixed", "descendants", "sse"])

    # shared knobs
    sp.add_argument("--c", type=int, default=200, help="concurrency")
    sp.add_argument("--n", type=int, default=20000, help="request count (when seconds=0)")
    sp.add_argument("--seconds", type=int, default=0, help="run duration seconds (0 means use --n)")

    # emit knobs
    sp.add_argument("--type", default="bench", help="event type for emit")
    sp.add_argument("--payload-bytes", type=int, default=128, help="approx payload size in bytes")
    sp.add_argument("--parents-mode", choices=["chain", "fork", "merge", "random"], default="chain")
    sp.add_argument("--parents-k", type=int, default=1, help="number of parents to attach (ignored in chain)")

    # read knobs
    sp.add_argument("--read-op", choices=["event", "children", "heads", "descendants"], default="event")
    sp.add_argument("--warm-depth", type=int, default=2, help="warm up depth for read pool")
    sp.add_argument("--read-ops", default="event,children,heads", help="mixed read ops, comma-separated")

    # mixed
    sp.add_argument("--write-ratio", type=float, default=0.2, help="mixed: probability of write (0-1)")

    # descendants scenario: build tree first
    sp.add_argument("--build-n", type=int, default=5000, help="descendants: number of emits to build a subtree")
    sp.add_argument("--build-c", type=int, default=200, help="descendants: build concurrency")

    # sse scenario
    sp.add_argument("--subs", type=int, default=200, help="sse: number of subscribers")
    sp.add_argument("--emit-rate", type=int, default=200, help="sse: emits per second during test")

    return p


async def main_async():
    parser = build_parser()
    args = parser.parse_args()

    if args.cmd != "scenario":
        raise RuntimeError("unknown command")

    if args.name == "emit":
        if args.seconds > 0:
            # 把 seconds 模式转换为按时间发请求（简单实现：循环发）
            # 这里给一个“持续 emit”的版本
            ct = CTClient(args.base, timeout=args.timeout)
            try:
                await ct.healthz()
                heads = await ensure_genesis_heads(ct)
                pool: List[int] = heads[:]
                sem = asyncio.Semaphore(args.c)
                stats = Stats()
                start = time.perf_counter()
                end_at = start + args.seconds
                i = 0

                async def one(i: int):
                    nonlocal pool
                    async with sem:
                        t0 = now_ms()
                        try:
                            parents = pick_parents(args.parents_mode, pool, args.parents_k)
                            payload = make_payload(args.payload_bytes)
                            meta = {"bench": "emit-seconds", "i": i}
                            out = await ct.emit(args.type, parents, payload, meta)
                            pool.append(int(out["event"]["id"]))
                            stats.add_ok(now_ms() - t0)
                        except Exception:
                            stats.add_fail(now_ms() - t0)

                tasks: List[asyncio.Task] = []
                while time.perf_counter() < end_at:
                    i += 1
                    tasks.append(asyncio.create_task(one(i)))
                    if len(tasks) >= args.c * 6:
                        await asyncio.gather(*tasks)
                        tasks.clear()
                if tasks:
                    await asyncio.gather(*tasks)

                elapsed = time.perf_counter() - start
                print("[emit-seconds]", stats.summary(elapsed))
                print(f"[emit-seconds] pool_size={len(pool)}")
            finally:
                await ct.close()
            return

        await scenario_emit(args)
        return

    if args.name == "read":
        if args.seconds <= 0 and args.n <= 0:
            raise ValueError("read: need --seconds>0 or --n>0")
        await scenario_read(args)
        return

    if args.name == "mixed":
        if args.seconds <= 0:
            # mixed 场景强烈建议用 seconds
            args.seconds = 10
        await scenario_mixed(args)
        return

    if args.name == "descendants":
        if args.seconds <= 0:
            args.seconds = 10
        await scenario_descendants(args)
        return

    if args.name == "sse":
        if args.seconds <= 0:
            args.seconds = 10
        await scenario_sse(args)
        return

    raise RuntimeError("unknown scenario name")


def main():
    # Windows 兼容
    if os.name == "nt":
        asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())  # type: ignore
    asyncio.run(main_async())


if __name__ == "__main__":
    main()
