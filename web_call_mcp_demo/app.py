from flask import Flask, request, jsonify
from flask_cors import CORS
#from langchain.chains import ConversationChain
#from langchain.memory import ConversationBufferMemory
#from langchain_openai import ChatOpenAI
from mcp import ClientSession
from mcp.client.streamable_http import streamablehttp_client
from openai import OpenAI
from dotenv import load_dotenv
import os
import sys
import asyncio
import json
import logging

load_dotenv()

app = Flask(__name__)
CORS(app)  #解决跨域问题

#配置控制台日志
def setup_console_logging():
    #创建控制台处理器(输出到 sys.stderr)
    console_handler = logging.StreamHandler(stream=sys.stderr)

    #设置日志格式
    formatter = logging.Formatter(
            '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
            )

    console_handler.setFormatter(formatter)

    console_handler.setLevel(logging.DEBUG)

    #添加 Flask 的logger
    app.logger.addHandler(console_handler)
    app.logger.setLevel(logging.DEBUG)

#初始化日志
setup_console_logging()

#
def convert_tool(tool):
    # 将MCP 工具转为兼容格式
    return {
           "type": "function",
           "function": {
           "name": tool.name,
           "description": tool.description,
           "parameters": tool.inputSchema                                                                       }
    }


#get mcp tools
async def get_mcp_tools():
    async with streamablehttp_client(url=os.getenv("MCP_SERVER_URL")) as (read,write,get_session_id):
        async with ClientSession(read,write) as session:
            # 初始化会话(必须步骤)
            await session.initialize()
            # 获取工具列表
            response = await session.list_tools()
            tools = [convert_tool(t) for t in response.tools]
            return tools

async def call_mcp_tool(tool_name,tool_args):
    # 执行工具
    async with streamablehttp_client(url=os.getenv("MCP_SERVER_URL")) as (read,write,get_session_id):
        async with ClientSession(read,write) as session:
            # 初始化会话(必须步骤)
            await session.initialize()
            result = await session.call_tool(tool_name, tool_args)
            print(f"\n\n[Calling tool {tool_name} with args {tool_args}\n\n")
    return result.content[0].text

# 初始化大模型的对话链
client = OpenAI(
        base_url = os.getenv("BASE_URL"),
        api_key = os.getenv("OPENAI_API_KEY"),
        )

@app.route('/chat', methods=['POST'])
def chat():
    user_input = request.json.get('message')
    if not user_input:
        return jsonify({"error": "No message provided"})

    messages = [
                {"role": "system", "content": "你是智能助手，帮助用户回答问题。"},
                {"role": "user", "content": user_input}
               ]
    # 调用 OpenAI API
    tools = asyncio.run(get_mcp_tools())
    response = client.chat.completions.create(
            model = os.getenv("MODEL"),
            messages = messages,
            tools = tools
            )

    #处理返回的内容
    content = response.choices[0]
    if content.finish_reason == "tool_calls":
        # 如是需要使用工具，就解析工具

        tool_call = content.message.tool_calls[0]
        tool_name = tool_call.function.name
        tool_args = json.loads(tool_call.function.arguments)

        tool_call_result = asyncio.run(call_mcp_tool(tool_name,tool_args))

        # 将模型返回的调用哪个工具数据和工具执行后的数据都存入message 中
        messages.append(content.message.model_dump())
        messages.append({
            "role": "tool",
            "content": tool_call_result,
            "tool_call_id": tool_call.id,
            })

        # 将上面的结果再返回给大模型用于生成最终结果
        response = client.chat.completions.create(
                model = os.getenv("MODEL"),
                messages = messages,
                )
    result = response.choices[0].message.content

    return jsonify({
        "response": result
        })


if __name__ == '__main__':
    app.run(host='0.0.0.0',port=5000,debug=True)

