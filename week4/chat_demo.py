from typing import Annotated
from typing_extensions import TypedDict
from langgraph.graph import StateGraph, START, END
from langgraph.graph.message import add_messages
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    model  = "qwen2.5",
    api_key = "ollama",
    #base_url = "http://172.29.20.187:11434/v1",
    base_url = "http://192.168.1.111:11434/v1",
)

class State(TypedDict):
    messages: Annotated[list, add_messages]

def chatbot(state: State):
    return {"messages": [llm.invoke(state["messages"])]}

graph_builder = StateGraph(State)
graph_builder.add_node("chatbot",chatbot)
graph_builder.add_edge(START,"chatbot")
graph_builder.add_edge("chatbot",END)
graph = graph_builder.compile()

def stream_graph_updates(user_input: str):
    for event in graph.stream({"messages": ["user",user_input]}):
        for value in event.values():
            print("Assistant:", value["messages"][-1].content)

while True:
    try:
        user_input = input("User: ")
        if user_input.lower() in ["quit","exit","q"]:
            print("Goodbye!")
            break
        # if user_input.lower() == "":
        #     continue
        stream_graph_updates(user_input)

    except:
        user_input = "What do you know about LangGraph ?"
        print("User: " + user_input)
        stream_graph_updates(user_input)
        break
