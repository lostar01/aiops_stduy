from openai import OpenAI

client = OpenAI(
    api_key="ollama",
    base_url="http://192.168.1.111:11434/v1",
)

completion = client.chat.completions.create(
    model="qwen2.5",
    messages=[
        {
            "role": "system",
            "content": "将以下英文翻译成中文",
        },
        {
            "role": "user",
            "content": "Thoughts on a Quiet Night\nThe moonlight shines before my bed, like frost upon the ground.\nI look up to see the moon, then lower my gaze, overwhelmed by homesickness.",
        },
    ],
)

print(completion.choices[0].message.content)
