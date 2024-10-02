from openai import OpenAI

client = OpenAI(
  api_key = "ollama",
  base_url = "http://192.168.1.108:11434/v1"
)

completion = client.chat.completions.create(
        #model="gpt-4o-mini",
        model="llama3",
        messages=[
            {
                "role": "system",
                "content": "你现在是一个云原生开发专家，你的工作是帮助用户解决技术问题。",
            },
            {
                "role": "user",
                "content": "kubernetes 的边缘网关架构怎么设计？请一步步给出说明"
            },
        ],
        )

print(completion.choices[0].message.content)
