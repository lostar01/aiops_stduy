import getpass
import os

# 定义函数，设置为定义环境变量
def __set_if_undefined(var: str):
    if not os.environ.get(var):
        os.environ[var] = getpass.getpass(f"Please provide you {var}:")


# 设置 OPENAI API 密钥
__set_if_undefined("OPEN_API_KEY")
__set_if_undefined("TAVILY_API_KEY")

#
import time
from langchain_community.tools.tavily_search import TavilySearchResults
from langchain_experimental.tools import PythonAstREPLTool

# 初始哈两个工具
tavily_tool = TavilySearchResults(max_result=5)

#执行python 代码工具，谨慎使用
python_repl_tool = PythonAstREPLTool()

#定义消息 Agent 的节点
from langchain_core.messages import HumanMessage

# 定义一个消息节点，把结果封装 HumanMessage 类型
def agent_node(state, agent, name):
    result = agent.invoke(state)
    return {
        "message": [HumanMessage(content=result["messages"][1].content, name=name)]
    }

# 定义 supervisor Agent 和调度逻辑
# 管理员 Agent， 负责决定下一个由那个Agent 来执行任务
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from langchain_openai import ChatOpenAI
from pydantic import BaseModel
from typing import Literal,Sequence
import operator
from typing_extensions import TypedDict,Annotated

from langchain_core.messages import BaseMessage

# 定义成员 Agent 和系统提示语，告诉 supervisor 要负责调度多个Agent
members = ["Researcher","Coder","AutoFixK8s","HumanHelp"]
system_prompt = (
    "You are a supervisor tasked with managing a conversation between the"
    " following workers: {members}. Given the following user request,"
    " respond with the worker to act next. Each worker will perform a"
    " task and respond with their results and status. When finished,"
    " respond with FINISH."
    "补充信息： Researcher 负责网络搜索； Coder 负责代码执行；AutoFixK8s 负责修复Kubernetes 问题; HumanHelp 寻求人工服务"
)

options = ["FINISH"] + members

# 定义 supervisor 的响应类， 选择一下执行的 Agent
class routeResponse(BaseModel):
    next: Literal[*options]

# 创建提示语模板
prompt = ChatPromptTemplate.from_messages(
    [
        ("system", system_prompt),
        MessagesPlaceholder(variable_name="messages"),
        (
            "system",
            "Given the conversation above, who should act next ?"
            "Or should we FINISH Select one of: {options}"
        ),
    ]
).partial(options=str(options),members=",".join(members))

class AgentState(TypedDict):
    messages: Annotated[Sequence[BaseMessage], operator.add]
    next: str


# 定义 LLM 模型和 supervisor_agent 函数
# llm = ChatOpenAI(model="gpt-4o")
llm = ChatOpenAI(model="qwen2.5",api_key="ollama",base_url="http://172.29.20.187:11434/v1")

def suprvisor_agent(state):
    supervisor_chain = prompt | llm.with_structured_output(routeResponse)
    return supervisor_chain.invoke(state)

# 定义K8S 自动修复工具
# 这里定义个 K8S 自动修复工具, 使用 OPENAI 模型生成 patch json
from langchain_core.tools import tool
from openai import OpenAI
from kubernetes import client, config, watch
import yaml

config.load_kube_config()
k8s_apps_v1 = client.AppsV1Api()

@tool
def auto_fix_k8s(deployment_name, namespace, event: str):
    """自动修复 K8S 问题"""
    # 先根据 deployment_name 获取 Deployment YAML
    # print("log:----", deployment_name, namespace, event)
    deployment = k8s_apps_v1.read_namespaced_deployment(
        name=deployment_name, namespace=namespace
    )
    deployment_dict = deployment.to_dict()
    # 移除不需要的字段
    deployment_dict.pop("status", None)
 

    if "metadata" in deployment_dict:
        deployment_dict["metadata"].pop("managed_fields",None)

    

    # 请求OpenAI 生成修复的 Patch JSON
    deployment_yaml = yaml.dump(deployment_dict)
    OpenAIClient = OpenAI(api_key="ollama",base_url="http://172.29.20.187:11434/v1")
    print("OpenAIClient: ", OpenAIClient)
    try:
        response = OpenAIClient.chat.completions.create(
            # model="gpt-4o",
            model="qwen2.5",
            response_format={"type": "json_object"},
            messages=[
                {
                    "role": "system",
                    "content": "你是一个助理用来输出 JSON"
                },
                {
                    "role": "user",
                    "content": f"""你现在是一个云原生技术专家，现在你需要根据 K8s 的报错生成一个能通过 kubectl patch 的一段 JSON 内容来修复问题。
            K8s 抛出的错误信息是：{event}
            工作负载的 YAML 是：
            {deployment_yaml}
        你生成的 patch JSON 应该可以直接通过 kubectl patch 命令使用，除此之外不要提供其他无用的建议，直接返回 JSON，且不要把 JSON 放在代码块内
        """,
                }
            ]
        )

        response1 = OpenAIClient.chat.completions.create(
            # model="gpt-4o",
            model="qwen2.5",
            response_format={"type": "json_object"},
            messages=[
                {
                    "role": "system",
                    "content": "你是一个助理用来输出 JSON"
                },
                {
                    "role": "user",
                    "content": f"""你现在是一个云原生技术专家，现在你需要验证 Kubernetes Patch Json: {response.choices[0].message.content} 是否正确，如果不正确请更正。
                 工作负载的 YAML 是：
                 {deployment_yaml}
                你重新生成的 patch JSON 应该可以直接通过 kubectl patch 命令使用，除此之外不要提供其他无用的建议，直接返回 JSON，且不要把 JSON 放在代码块内  
        """,
                }
            ]
        )
        # print("===============\n",response.choices[0].message.content,"\n==============")

        json_opt = response1.choices[0].message.content
        print(json_opt)
    except Exception as e:
        print(f"生成Patch JSON失败： {str(e)}")
        
    # Apply Patch JSON
    try:
        k8s_apps_v1.patch_namespaced_deployment(
            name=deployment_name,
            namespace=namespace,
            body=yaml.safe_load(json_opt)
        )
    except Exception as e:
        print(f"修复失败： {str(e)}")
        return f"修复失败： {str(e)}"
    
    return f"工作负载自动修复成功！"


# 定义人工帮助的Tool
# 用于在无法自动修复时候发送飞书消息通知

import requests
import json

@tool
def human_help(event_message: str):
    """无法修复问题时候寻求人工帮助"""
    url = "https://open.feishu.cn/open-apis/bot/v2/hook/d219b128-1520-40b3-b7ce-8df8c01422d2"
    headers = {"Content-Type": "application/json"}
    data = {"msg_type": "text","content": {"text": event_message}}
    response = requests.post(url, headers=headers, data=json.dumps(data))
    return response.status_code


# 定义工作流和 Graph （有向有环图）
# 创建工作流， 并且给节点添加 route 路由逻辑

from langgraph.graph import END, StateGraph, START
from langgraph.prebuilt import create_react_agent
import functools

# 创建Agent
research_agent = create_react_agent(llm, tools=[tavily_tool])
research_node = functools.partial(agent_node, agent=research_agent, name="Researcher")

# Code Agent
code_agent = create_react_agent(llm, tools=[python_repl_tool])
code_node = functools.partial(agent_node, agent=code_agent, name="Coder")

# Auto fix Agent
auto_fix_agent = create_react_agent(llm, tools=[auto_fix_k8s])
auto_fix_node = functools.partial(agent_node, agent=auto_fix_agent, name="AutoFixK8S")

# human help agent
human_help_agent = create_react_agent(llm, tools=[human_help])
human_help_node = functools.partial(agent_node, agent=human_help_agent, name="HumanHelp")

# 创建Graph 并添加节点
workflow = StateGraph(AgentState)
workflow.add_node("Researcher", research_node)
workflow.add_node("Coder", code_node)
workflow.add_node("supervisor", suprvisor_agent)
workflow.add_node("AutoFixK8s", auto_fix_node)
workflow.add_node("HumanHelp", human_help_node)

# 定义路由逻辑
for member in members:
    workflow.add_edge(member, "supervisor")

conditional_map = { k: k  for k in members}
conditional_map["FINISH"] = END
workflow.add_conditional_edges("supervisor", lambda x: x["next"], conditional_map)

workflow.add_edge(START, "supervisor")

# 编译 Graph
graph = workflow.compile()

# # 展示 Graph
# from IPython.display import display, Image
# try:
#     display(Image(graph.get_graph().draw_mermaid_png()))
# except Exception as e:
#     pass


# 硬编码，测试效果
for s in graph.stream(
    {
        "messages": [
            HumanMessage(content="deployment: nginx, default,  event: Back-off pulling image 'nginx:latess'")
        ]
    }
):
    if "__end__" not in s:
        print(s)
        print("--- ---")


