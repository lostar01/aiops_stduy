from openai import OpenAI

client = OpenAI(
        api_key = "ollama",
        base_url = "http://192.168.1.107:11434/v1"
        )

completion = client.chat.completions.create(
        model = "llama3.1",
        response_format={"type": "json_object"},
        messages = [
            {
                "role": "system",
                "content": '你现在是一个JSON 对象提取专家,请参考我的JSON 定义输出 JSON 对象。示例: {"service_name": "", "action": ""},其中，action 可以是 get_log (获取日志) 、restart(重启服务) 、delete (删除工作负载)',

            },
            {   "role": "user",
                "content": "帮我重启 payment 服务。",
            },
            ],
        )

print(completion.choices[0].message.content)
