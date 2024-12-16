from langchain.text_splitter import MarkdownHeaderTextSplitter
# from langchain_openai import OpenAIEmbeddings
from langchain_ollama import OllamaEmbeddings
from langchain_community.vectorstores.chroma import Chroma
from langchain_core.prompts import PromptTemplate
from langchain_openai import ChatOpenAI
# from langchain.schema.runnable import RunnablePassthrough
# from langchain.schema import StrOutputParser
from string import Template

import os
import uuid


file_path = os.path.join('data','data.md')

with open(file_path,'r',encoding='utf-8') as file:
    docs_string = file.read()



headers_to_split_on = [
    ("#","Header 1"),
    ("##","Header 2"),
    ("###","Header 3"),
]

text_spliter = MarkdownHeaderTextSplitter(headers_to_split_on=headers_to_split_on)
splits = text_spliter.split_text(docs_string)
print("Length of splits: " + str(len(splits)))

# print(splits)

# 将运维知识的每一块文本向量化( Embedding)
# 保存到随机目录里

random_directory = "./" + str(uuid.uuid4)
# embedding = OllamaEmbeddings(model="bge-m3",base_url="http://192.168.1.111:11434")
embedding = OllamaEmbeddings(model="nomic-embed-text",base_url="http://192.168.1.111:11434")
vectorstore = Chroma.from_documents(documents=splits, embedding=embedding, persist_directory=random_directory)
vectorstore.persist()

# 传统 RAG 流程
retriever = vectorstore.as_retriever()

# 提示语模板
template = """使用上下文来回答最后的问题。
如果你不知道答案，就说不知道，不要试图编造答案。
最多使用三句话，并尽可能简洁回答。
在答案的最后一定要说“谢谢询问！”。

{context}

Question: {question}

Helpful Answer: 
"""

custom_rag_prompt = PromptTemplate.from_template(template)

def format_docs(docs):
    print("匹配到的运维知识库片段：\n", "\n\n".join(doc.page_content for doc in docs))
    return "\n\n".join(doc.page_content for doc in docs)

llm = ChatOpenAI(model="qwen2.5",api_key="ollama",base_url="http://192.168.1.111:11434/v1")

# rag_chain = (
#     {"context": retriever | format_docs, "question": RunnablePassthrough()}
#     | custom_rag_prompt
#     | llm
#     | StrOutputParser()
# )

# # 传统 RAG 无法回答的问题
# res = rag_chain.invoke("谁管的系统最多？")
# print("\n\nLLM 回答：", res)

# Agent RAG 流程
# 文本相似性检索
def search_docs(query, k=1):
    results = vectorstore.similarity_search_with_score(
        query=query,
        k=k,
    )
    return "\n\n".join(doc.page_content for doc, score in results)

user_query = "谁管的系统最多? "

check_can_answer_system_prompt = """
根据上下文识别是否能够回答问题，如果不能，则返回 JSON 字符串 "{"can_answer": false}"，如果可以则返回 "{"can_answer": true}"。
上下文：\n $context
问题：$question
"""

k = 1
docs = ""
while True:
    # 通过检索找到相关文档，每次循环增加一个检索文档数量，最大 15 个文档块
    print("第", k , "次检索")
    if k > 15:
        break
    docs = search_docs(user_query,k)
    print("匹配到的文档块", docs)
    template = Template(check_can_answer_system_prompt)
    filled_prompt = template.substitute(question=user_query, context=docs)
    print("show debug ----", filled_prompt,"\n")
    # 检查上下文是否能够回答问题
    messages = [
        (
            "system",
            filled_prompt,
        ),
        ("human", "开始检查上下文是否足够回答问题。"),
    ]

    llm_message = llm.invoke(messages)
    content = llm_message.content
    print("\nLLM Res: ", content, "\n")
    if content == '{"can_answer": true}':
        break
    else:
        k += 1

print("匹配到能够回答问题的知识库，开始进行回答\n")


# 最终推理
final_system_prompt = """
您是问答任务的助手，使用检索到的上下文来回答用户提出的问题。如果你不知道答案，就说不知道。最多使用三句话并保持答案简洁。
"""

final_messages = [
    (
        "system",
        final_system_prompt,
    ),
    ("human", "上下文：\n"+ docs +"\n问题：" + user_query),
]
llm_message = llm.invoke(final_messages)
content = llm_message.content
print("\nLLM Final Res: ", content, "\n")