# _*_ coding: utf-8 _*_

import argparse
import json
import statistics
import time
from typing import Any

import requests
from celestialflow import TaskManager


def now_ms() -> float:
    return time.perf_counter() * 1000.0

session = requests.Session()

# ===============================
# 单个 Emit 任务的执行函数
# ===============================

def emit_once(base_url, payload_txt, _) -> float:
    """
    执行一次 /emit 请求
    返回：本次请求的 latency（ms）
    """
    url = f"{base_url}/emit"
    payload = {
        "type": "bench",
        "parents": [],
        "payload": {
            "v": payload_txt
        },
        "meta": {}
    }

    t0 = now_ms()
    r = session.post(url, json=payload, timeout=10)
    t1 = now_ms()

    if r.status_code != 200:
        raise RuntimeError(f"HTTP {r.status_code}")

    return t1 - t0


# ===============================
# TaskManager：Emit Bench
# ===============================

class EmitBenchManager(TaskManager):
    def process_result_dict(self):
        results_list = []

        for result in self.get_success_dict().values():
            results_list.append(result)

        return results_list


# ===============================
# 主入口
# ===============================

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--base", default="http://127.0.0.1:7777")
    ap.add_argument("--n", type=int, default=10000)
    ap.add_argument("--c", type=int, default=20)
    ap.add_argument("--payload-bytes", type=int, default=32)
    args = ap.parse_args()

    # ---- 构造 payload ----
    payload_str = "x" * args.payload_bytes

    # ---- 构造任务列表 ----
    task_list = [
        (args.base, payload_str, index)
        for index in range(args.n)
    ]

    # ---- TaskManager ----
    manager = EmitBenchManager(
        emit_once,
        execution_mode="thread",
        worker_limit=args.c,
        unpack_task_args=True,
        enable_result_cache=True,
        enable_duplicate_check=False,
        show_progress=True,
        progress_desc="emit-bench",
    )

    start = time.perf_counter()
    manager.start(task_list)
    elapsed = time.perf_counter() - start
    results = manager.process_result_dict()

    # ---- 统计 ----
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


if __name__ == "__main__":
    main()
