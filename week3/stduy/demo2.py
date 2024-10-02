from openai import OpenAI

client = OpenAI(
    api_key="ollama"
    base_url="https://api.apiyi.com/v1",
    )

messages = [
    {
        "role": "system",
        "content": "你现在是一个运维专家，你的工作是帮助用户解决技术问题",

    }
]

while True:
    user_input = input("输入问题：")
    messages.append(
        {
            "role": "user",
            "content": user_input,
        }
    )

    print("当前聊天上下文：",messages)

    completion = client.chat.completions.create(
        model = "gpt-4o-mini",
        messages = messages,
        )

    reply = completion.choices[0].message.content
    print(f"\n运维专家：{reply}")

    messages.append(
        {
            "role": "assistant",
            "content": reply,

        }
    )
