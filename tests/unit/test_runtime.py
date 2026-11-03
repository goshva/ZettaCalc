from src.harness.llm import LLMStep, ScriptedLLM
from src.harness.runtime import Budget, Tool, WritePolicy, run_agent
from src.harness.tools import make_tools


def test_final_result(catalog):
    llm = ScriptedLLM(
        [
            LLMStep(tool="get_catalog_item", args={"sku": "SKU-GROUT-5"}, tokens=100),
            LLMStep(final={"ok": True}, tokens=50),
        ]
    )
    trace = run_agent("t", Budget(), llm, make_tools(catalog))
    assert trace.stop_reason == "final" and trace.result == {"ok": True}
    assert trace.steps[0]["result"]["price_rub"] == 1150.0


def test_unknown_sku_is_reported_not_invented(catalog):
    llm = ScriptedLLM(
        [LLMStep(tool="get_catalog_item", args={"sku": "X"}), LLMStep(final={})]
    )
    trace = run_agent("t", Budget(), llm, make_tools(catalog))
    assert trace.steps[0]["result"] == {"found": False, "sku": "X"}


def test_step_budget_stops_agent(catalog):
    steps = [LLMStep(tool="get_catalog_item", args={"sku": f"S{i}"}) for i in range(10)]
    trace = run_agent("t", Budget(max_steps=3), ScriptedLLM(steps), make_tools(catalog))
    assert trace.stop_reason == "budget_steps" and len(trace.steps) == 3


def test_token_budget_stops_agent(catalog):
    steps = [LLMStep(tool="get_catalog_item", args={"sku": "A"}, tokens=900)] * 5
    trace = run_agent("t", Budget(max_tokens=1000), ScriptedLLM(steps), make_tools(catalog))
    assert trace.stop_reason == "budget_tokens"


def test_cost_budget_stops_agent(catalog):
    steps = [LLMStep(tool="get_catalog_item", args={"sku": str(i)}, cost_rub=30) for i in range(5)]
    trace = run_agent("t", Budget(max_cost_rub=50), ScriptedLLM(steps), make_tools(catalog))
    assert trace.stop_reason == "budget_cost"


def test_anti_loop(catalog):
    steps = [LLMStep(tool="get_catalog_item", args={"sku": "A"})] * 10
    trace = run_agent("t", Budget(), ScriptedLLM(steps), make_tools(catalog))
    assert trace.stop_reason == "anti_loop"


def test_write_tool_denied_without_approval():
    called = []
    tools = {"order": Tool(lambda: called.append(1) or {}, writes=True)}
    llm = ScriptedLLM([LLMStep(tool="order"), LLMStep(final={})])
    trace = run_agent("t", Budget(), llm, tools, WritePolicy(allow_write=True, approved=False))
    assert called == [] and trace.steps[0]["result"] == {"error": "write_denied"}


def test_write_tool_runs_with_approval():
    called = []
    tools = {"order": Tool(lambda: called.append(1) or {"ok": 1}, writes=True)}
    llm = ScriptedLLM([LLMStep(tool="order"), LLMStep(final={})])
    run_agent("t", Budget(), llm, tools, WritePolicy(allow_write=True, approved=True))
    assert called == [1]
