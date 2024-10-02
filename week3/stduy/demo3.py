from langchain.chat_models import ChatOpenAI
from langchain.chains import ConversationChain
from langchain.memory import ConversationBufferMemory

llm = ChatOpenAI(
    model = "llam3",
    openai_api_key="xxxx",
    openai_api_base="http://192.168.1.108:11434",
    max_tokens=None,
    timeout=None,
)

memory = ConversationBufferMemory()
conversation = ConversationChain(
      llm = llm,
      memory = memory,
      verbose = False,
    )

while True:
    user_input = input("输入问题：")
    reply = conversation.predict(input=user_input)
    print("当前聊天上下文：", conversation.memory)
    print(f"\n\n运维专家： {reply}")
