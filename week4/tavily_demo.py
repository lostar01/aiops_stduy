import os
from langchain_openai import ChatOpenAI
from typing import Literal
# from langchain_core.tools import tool
# from IPython.display import Image, display
from langgraph.prebuilt import ToolNode
from langgraph.graph import StateGraph, MessagesState
from langchain_community.tools.tavily_search import TavilySearchResults

# Tavily API Key
os.environ["TAVILY_API_KEY"] = "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"


def call_model(state: MessagesState):
    messages = state["messages"]
    response = model_with_tools.invoke(messages)
    return {"messages": [response]}

def should_continue(state: MessagesState) -> Literal["tools", "__end__"]:
    messages = state["messages"]
    last_message = messages[-1]
    if last_message.tool_calls:
        return "tools"
    return "__end__"

tools = [TavilySearchResults(max_results=1)]
model_with_tools = ChatOpenAI(model="qwen2.5", api_key="ollama", base_url="http://192.168.1.111:11434/v1", temperature=0).bind_tools(tools)
tool_node = ToolNode(tools)
workflow = StateGraph(MessagesState)
workflow.add_node("agent", call_model)
workflow.add_node("tools", tool_node)

workflow.add_edge("__start__", "agent")
workflow.add_conditional_edges(
    "agent",
    should_continue,
)
workflow.add_edge("tools", "agent")
app = workflow.compile()

# try:
#     display(Image(app.get_graph().draw_mermaid_png()))
# except Exception:
#     pass

for chunk in app.stream(
    {"messages": [("human", "2024年广州程序员平均薪酬")]}, stream_mode="values"
):
    chunk["messages"][-1].pretty_print()