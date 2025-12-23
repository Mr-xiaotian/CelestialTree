#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
Redis emit-style benchmark
- 单条写入
- 无 pipeline
- INCR + RPUSH
- payload 大小可控
"""

import argparse
import asyncio
import json
import os
import random
import statistics
import string
import time
from typing import List
from dotenv import load_dotenv

import redis.asyncio as redis


load_dotenv()
redis_host = os.getenv("REDIS_HOST")
redis_port = os.getenv("REDIS_PORT")
redis_possward = os.getenv("REDIS_PASSWORD")

def now_ms() -> float:
    return time.perf_counter() * 1000.0


def pct(values: List[float], p: float) -> float:
    if not values:
        return float("nan")
    v = sorted(values)
    k = (len(v) - 1) * (p / 100.0)
    f = int(k)
    c = min(f + 1, len(v) - 1)
    if f == c:
        return v[f]
    return v[f] + (v[c] - v[f]) * (k - f)


def make_payload(payload_bytes: int) -> str:
    if payload_bytes <= 0:
        return "hi"
    return "".join(random.choice(string.ascii_letters + string.digits)
                   for _ in range(payload_bytes))


async def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--redis", default=f"redis://default:{redis_possward}@{redis_host}:{redis_port}/0")
    ap.add_argument("--n", type=int, default=10000)
    ap.add_argument("--c", type=int, default=50)
    ap.add_argument("--payload-bytes", type=int, default=32)
    ap.add_argument("--key", default="ct:events")
    args = ap.parse_args()

    r = redis.from_url(args.redis, decode_responses=True)

    # 清理
    await r.delete(args.key, "ct:id")

    sem = asyncio.Semaphore(args.c)
    lat_ms: List[float] = []
    ok = 0
    fail = 0

    async def one(i: int):
        nonlocal ok, fail
        async with sem:
            t0 = now_ms()
            try:
                eid = await r.incr("ct:id")
                payload = make_payload(args.payload_bytes)
                ev = {
                    "id": eid,
                    "type": "bench",
                    "payload": payload,
                }
                await r.rpush(args.key, json.dumps(ev))
                lat_ms.append(now_ms() - t0)
                ok += 1
            except Exception:
                lat_ms.append(now_ms() - t0)
                fail += 1

    start = time.perf_counter()
    tasks = [asyncio.create_task(one(i)) for i in range(args.n)]
    await asyncio.gather(*tasks)
    elapsed = time.perf_counter() - start

    total = ok + fail
    rps = total / elapsed

    print(
        f"[redis] total={total} ok={ok} fail={fail} "
        f"rps={rps:.1f} "
        f"lat_ms(avg={statistics.mean(lat_ms):.2f} "
        f"p50={pct(lat_ms,50):.2f} "
        f"p90={pct(lat_ms,90):.2f} "
        f"p95={pct(lat_ms,95):.2f} "
        f"p99={pct(lat_ms,99):.2f} "
        f"max={max(lat_ms):.2f})"
    )

    await r.close()


if __name__ == "__main__":
    if os.name == "nt":
        asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())
    asyncio.run(main())
