"""Ansible callback plugin: one JSON object per line, emitted as events
happen — not the built-in `json` callback, which buffers everything and
dumps it once at the very end of the run. Converge needs real progress for
its run log, so this exists.

Enabled by setting, on the ansible-playbook subprocess:
    ANSIBLE_STDOUT_CALLBACK=converge_json
    ANSIBLE_CALLBACK_PLUGINS=<this file's directory>

Event shape (one per stdout line):
    {"event": "task_start", "task": "core : Install pixi global tools"}
    {"event": "result", "status": "ok|changed|failed|skipped|unreachable",
     "task": "...", "host": "...", "msg": "..."}
    {"event": "stats", "host": "...", "ok": N, "changed": N, "failed": N,
     "skipped": N, "unreachable": N}
    {"event": "play_start", "play": "..."}
"""
from __future__ import annotations

import json
import sys

from ansible.plugins.callback import CallbackBase

DOCUMENTATION = """
    name: converge_json
    type: stdout
    short_description: line-delimited JSON events for the Converge run log
"""


class CallbackModule(CallbackBase):
    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = "stdout"
    CALLBACK_NAME = "converge_json"

    def _emit(self, event: dict) -> None:
        print(json.dumps(event), flush=True)

    def v2_playbook_on_play_start(self, play):
        self._emit({"event": "play_start", "play": play.get_name() or "play"})

    def v2_playbook_on_task_start(self, task, is_conditional):
        self._emit({"event": "task_start", "task": task.get_name()})

    def _result(self, status, result):
        self._emit({
            "event": "result",
            "status": status,
            "task": result.task_name or result._task.get_name(),
            "host": result._host.get_name(),
            "msg": _short_msg(result._result),
        })

    def v2_runner_on_ok(self, result):
        status = "changed" if result._result.get("changed") else "ok"
        self._result(status, result)

    def v2_runner_on_failed(self, result, ignore_errors=False):
        self._result("failed", result)

    def v2_runner_on_skipped(self, result):
        self._result("skipped", result)

    def v2_runner_on_unreachable(self, result):
        self._result("unreachable", result)

    def v2_playbook_on_stats(self, stats):
        for host in stats.processed.keys():
            s = stats.summarize(host)
            self._emit({
                "event": "stats", "host": host,
                "ok": s["ok"], "changed": s["changed"], "failed": s["failures"],
                "skipped": s["skipped"], "unreachable": s["unreachable"],
            })


def _short_msg(result: dict) -> str:
    for key in ("msg", "stdout", "reason"):
        val = result.get(key)
        if val:
            return str(val).strip().splitlines()[0][:200]
    return ""


if __name__ == "__main__":
    sys.exit("this is an ansible callback plugin, not a script")
