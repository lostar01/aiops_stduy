from openai import OpenAI

import json
import time

client = OpenAI(
    api_key = "ollama",
    base_url = "http://192.168.1.107:11434/v1",
    )

def analyze_loki_log(query_str):
    print("\nParameters for function calls： ", query_str)
    return json.dumps({"log": "this is error log"})

def run_conversation():
    """Query to view logs of app=grafana and keywords containing Error"""

    # 步骤一: 把所有预定的 function 传给 chatgpt
    query = input("输入查询指令:")
    messages = [
        {
          "role": "system",
          "content": "You are a Loki log analysis assistant that can help users analyze Loki logs. You can call multiple functions to assist users in completing tasks and ultimately attempt to analyze the cause of errors"
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
              "name": "analyze_loki_log",
              "description": "Retrieve logs from Loki",
              "parameters": {
                 "type": "object",
                 "properties": {
                    "query_str": {
                      "type": "string",
                      "description": 'Loki query string, for example: {app="grafana"} |="Error"',
                    },
                  },
                 "required": ["query_str"]
               },
            },
        }
    ]

    response = client.chat.completions.create(
              model = "llama3.1",
              messages = messages,
              tools = tools,
              tool_choice = "auto",
            )

    response_message = response.choices[0].message
    tool_calls = response_message.tool_calls
    print("\nChatGPT want to call function:", tool_calls)

    # 步骤二：检查LLM 是否调用了 function
    if tool_calls is None:
        print("not tool_calls")
    else:
        available_functions = {
          "analyze_loki_log": analyze_loki_log,
        }

        messages.append(response_message)
        # 步骤三： 把每次 function 调用和返回信息给 model
        for tool_call in tool_calls:
            function_name = tool_call.function.name
            function_to_call = available_functions[function_name]
            function_args = json.loads(tool_call.function.arguments)
            function_response = function_to_call(**function_args)
            messages.append(
                {
                    "tool_call_id": tool_call.id,
                    "role": "tool",
                    "name": function_name,
                    "content": function_response,
                }
            )

            # 步骤四： 把 function calling 的结果传给 model, 进行对话
            response = client.chat.completions.create(
                model = "llama3.1",
                messages = messages,
                )
            return response.choices[0].message.content

print("LLM Res: ", run_conversation())
