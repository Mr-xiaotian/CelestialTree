# _*_ coding: utf-8 _*_

import os
import argparse
import base64
import statistics
import time

import requests, httpx
from celestialflow import TaskExecutor
from celestialtree import Client


def now_ms() -> float:
    return time.perf_counter() * 1000.0

CTREE_HOST = "127.0.0.1"
CTREE_PORT = 7777
CTREE_GRPC_PORT = 7778

session = requests.Session()
async_client = httpx.AsyncClient(trust_env=False)
ctree_client = Client(host=CTREE_HOST, port=CTREE_PORT, grpc_port=CTREE_GRPC_PORT)

# ===============================
# 单个 Emit 任务的执行函数
# ===============================

def emit_once(payload_size, _) -> float:
    """
    执行一次 /emit 请求
    返回：本次请求的 latency（ms）
    """
    payload_raw = os.urandom(payload_size)
    payload_b64 = base64.b64encode(payload_raw).decode("ascii")

    url = f"http://{CTREE_HOST}:{CTREE_PORT}/emit"
    payload = {
        "type": "bench",
        "parents": [],
        "message": f"bench payload {payload_size}B",
        "payload": payload_b64 
    }

    t0 = now_ms()
    r = session.post(url, json=payload, timeout=10)
    t1 = now_ms()

    if r.status_code != 200:
        raise RuntimeError(f"HTTP {r.status_code}")

    return t1 - t0


def emit_once_http(payload_size, _) -> float:
    """
    Docstring for emit_once_http
    
    :param payload_size: Description
    :param _: Description
    :return: Description
    :rtype: float
    """
    payload_raw = os.urandom(payload_size)
    payload_b64 = base64.b64encode(payload_raw).decode("ascii")

    payload = {
        "payload_b64": payload_b64,
    }

    t0 = now_ms()
    ctree_client.emit("bench", [], f"bench payload {payload_size}B", payload)
    t1 = now_ms()

    return t1 - t0


def emit_once_grpc(payload_size, _) -> float:
    """
    Docstring for emit_once_grpc
    
    :param payload_size: Description
    :param _: Description
    :return: Description
    :rtype: float
    """
    payload_raw = os.urandom(payload_size)
    payload_b64 = base64.b64encode(payload_raw).decode("ascii")

    payload = {
        "payload_b64": payload_b64,
    }

    t0 = now_ms()
    ctree_client.emit_grpc("bench", [], f"bench payload {payload_size}B", payload)
    t1 = now_ms()

    return t1 - t0


async def emit_once_async(
    payload_size: int,
    _
) -> float:
    """
    执行一次 /emit 请求（async）
    返回：本次请求的 latency（ms）
    """
    payload_raw = os.urandom(payload_size)
    payload_b64 = base64.b64encode(payload_raw).decode("ascii")

    url = f"http://{CTREE_HOST}:{CTREE_PORT}/emit"
    payload = {
        "type": "bench",
        "parents": [],
        "message": f"bench payload {payload_size}B",
        "payload": payload_b64,
    }

    t0 = now_ms()
    r = await async_client.post(url, json=payload)
    t1 = now_ms()

    if r.status_code != 200:
        raise RuntimeError(f"HTTP {r.status_code}")

    return t1 - t0


# ===============================
# TaskExecutor：Emit Bench
# ===============================

class EmitBenchExecutor(TaskExecutor):
    def process_result_dict(self):
        results_list = []

        for result in self.get_success_dict().values():
            results_list.append(result)

        return results_list


# ===============================
# 主入口
# ===============================

def print_stats(args, results, elapsed):
    lat_ms = [r for r in results if isinstance(r, (int, float))]
    ok = len(lat_ms)
    fail = args.n - ok
    rps = ok / elapsed if elapsed > 0 else 0.0

    lat_ms.sort()

    def pct(p: float) -> float:
        if not lat_ms:
            return 0.0
        idx = int((len(lat_ms) - 1) * p)
        return lat_ms[idx]

    print(
        f"[ct-taskflow] total={args.n} ok={ok} fail={fail} "
        f"rps={rps:.1f} "
        f"lat_ms(avg={statistics.mean(lat_ms):.2f} "
        f"p50={pct(0.50):.2f} "
        f"p90={pct(0.90):.2f} "
        f"p99={pct(0.99):.2f} "
        f"max={max(lat_ms):.2f})"
    )


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--n", type=int, default=10000)
    ap.add_argument("--c", type=int, default=20)
    ap.add_argument("--payload-bytes", type=int, default=32)
    args = ap.parse_args()

    # ---- 构造任务列表 ----
    task_list = [
        (args.payload_bytes, index)
        for index in range(args.n)
    ]

    # ---- TaskExecutor ----
    executor = EmitBenchExecutor(
        emit_once,
        execution_mode="thread",
        worker_limit=args.c,
        unpack_task_args=True,
        enable_success_cache=True,
        enable_duplicate_check=False,
        show_progress=True,
        progress_desc="emit-bench",
    )

    start = time.perf_counter()
    executor.start(task_list)
    elapsed = time.perf_counter() - start
    results = executor.process_result_dict()

    # ---- 统计 ----
    print_stats(args, results, elapsed)

def main_http():
    ap = argparse.ArgumentParser()
    ap.add_argument("--n", type=int, default=10000)
    ap.add_argument("--c", type=int, default=20)
    ap.add_argument("--payload-bytes", type=int, default=32)
    args = ap.parse_args()

    # ---- 构造任务列表 ----
    task_list = [
        (args.payload_bytes, index)
        for index in range(args.n)
    ]

    # ---- TaskExecutor ----
    executor = EmitBenchExecutor(
        emit_once_http,
        execution_mode="serial",
        worker_limit=args.c,
        unpack_task_args=True,
        enable_success_cache=True,
        enable_duplicate_check=False,
        show_progress=True,
        progress_desc="emit-bench",
    )

    start = time.perf_counter()
    executor.start(task_list)
    elapsed = time.perf_counter() - start
    results = executor.process_result_dict()

    # ---- 统计 ----
    print_stats(args, results, elapsed)

def main_grpc():
    ap = argparse.ArgumentParser()
    ap.add_argument("--n", type=int, default=10000)
    ap.add_argument("--c", type=int, default=20)
    ap.add_argument("--payload-bytes", type=int, default=32)
    args = ap.parse_args()

    # ---- 构造任务列表 ----
    task_list = [
        (args.payload_bytes, index)
        for index in range(args.n)
    ]

    # ---- TaskExecutor ----
    executor = EmitBenchExecutor(
        emit_once_grpc,
        execution_mode="serial",
        worker_limit=args.c,
        unpack_task_args=True,
        enable_success_cache=True,
        enable_duplicate_check=False,
        show_progress=True,
        progress_desc="emit-bench",
    )

    start = time.perf_counter()
    executor.start(task_list)
    elapsed = time.perf_counter() - start
    results = executor.process_result_dict()

    # ---- 统计 ----
    print_stats(args, results, elapsed)

async def main_async():
    ap = argparse.ArgumentParser()
    ap.add_argument("--n", type=int, default=10000)
    ap.add_argument("--c", type=int, default=20)
    ap.add_argument("--payload-bytes", type=int, default=32)
    args = ap.parse_args()

    # ---- 构造任务列表 ----
    task_list = [
        (args.payload_bytes, index)
        for index in range(args.n)
    ]

    executor = EmitBenchExecutor(
        emit_once_async,               # ⚠️ async 版 emit
        execution_mode="async",
        worker_limit=args.c,
        unpack_task_args=True,
        enable_success_cache=True,
        enable_duplicate_check=False,
        show_progress=True,
        progress_desc="emit-bench",
    )

    start = time.perf_counter()
    await executor.start_async(task_list)   # ✅ await，而不是 create_task
    elapsed = time.perf_counter() - start

    results = executor.process_result_dict()
    print_stats(args, results, elapsed)


if __name__ == "__main__":
    main_grpc()
    # asyncio.run(main_async())
    pass
