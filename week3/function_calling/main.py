from openai import OpenAI
import json
import time

client = OpenAI(
    api_key = "ollama",
    base_url = "http://172.29.20.187:11434/v1"
    )

def modify_config(service_name,key,value):
    print("\n 函数调用的参数：", service_name,key,value)
    return json.dumps({"log": "service={} key={} value={}".format(service_name,key,value)})

def restart_service(service_name):
    print("\n 准备重启 ", service_name)
    return json.dumps({"log": "重启{}成功".format(service_name)})

def apply_manifest(resource_type,image):
    print("\n 应用manifest", resource_type,image)
    return json.dumps({"log":"apply manifest done"})

def run_conversation():
    """ Action: 定义运维操作"""

    # 步骤一： 把所有预定义的function 传给 chatgpt
    query = input("输入指令:")

    messages = [
        {
            "role": "system",
            "content": "你是一个运维助手,你可以帮助用户执行运维操作，你可以调用多个函数来帮助用户完成任务"
        },
        {
            "role": "user",
            "content": query,
        },
        ]
    tools = [
        {
            "type": "function",
            "function": {
                "name": "modify_config",
                "description": "修改配置",
                "parameters": {
                    "service_name": {
                        "type": "string",
                        "description": "服务名"
                    },
                    "key": {
                        "type": "string",
                        "description": "主键key"
                    },
                    "value": {
                        "type": "string",
                        "description": "值为value"
                    },
                    "required": ["service_name","key","value"],
                },
            },
        },

        {
            "type": "function",
            "function": {
                "name": "restart_service",
                "description": "重启服务",
                "parameters": {
                    "service_name": {
                        "type": "string",
                        "description": "服务名"
                    },
                    "required": ["service_name"],
                },
            },
        },

        {
            "type": "function",
            "function": {
                "name": "apply_manifest",
                "description": "部署资源apply manifest",
                "parameters": {
                    "resource_type": {
                        "type": "string",
                        "description": "资源类型"
                    },
                    "image": {
                        "type": "string",
                        "description": "镜像名称"
                    },
                    "required": ["resource_type","image"],
                },
            },
        },
        ]

    response = client.chat.completions.create(
       #model = "llama3.1",
       model = "qwen2.5",
       messages = messages,
       tools = tools,
       tool_choice="auto",
       )

    response_message = response.choices[0].message
    tool_calls = response_message.tool_calls
    print("\n ChatGPT want to call function：", tool_calls)

    # 步骤二： 检查 LLM 是否调用了 function
    if tool_calls is None:
        print("not tool_calls")
    if tool_calls:
        available_functions = {
            "modify_config": modify_config,
            "restart_service": restart_service,
            "apply_manifest": apply_manifest
        }

        messages.append(response_message)

        # 步骤三： 把每次 function 调用和返回的信息传给 model
        for tool_call in tool_calls:
            function_name = tool_call.function.name
            function_to_call = available_functions[function_name]
            function_args = json.loads(tool_call.function.arguments)
            print(function_args)
            function_response = function_to_call(**function_args)

            messages.append(
                {
                    "tool_call_id": tool_call.id,
                    "role": "tool",
                    "name": function_name,
                    "content": function_response,
                }
                )

            # 步骤四: 把function calling 的结果给model, 进行对话
            response = client.chat.completions.create(
                model = "llama3.1",
                messages = messages,
            )

            return response.choices[0].message.content

print("LLM Res：", run_conversation())
