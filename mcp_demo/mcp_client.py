import asyncio
import os
import json
from openai import OpenAI
from dotenv import load_dotenv
from mcp import ClientSession
from mcp.client.streamable_http import streamablehttp_client
from contextlib import AsyncExitStack

#加载 .env 文件，确保 API KEY 受到保护
load_dotenv()

class MCPClient:
    def __init__(self):
        """初始化 MCP 客户端"""
        self.tools = []
        self.exit_stack = AsyncExitStack()
        self.mcp_server_url = os.getenv("MCP_SERVER_URL")
        self.openai_api_key = os.getenv("OPENAI_API_KEY")
        self.base_url = os.getenv("BASE_URL")
        self.model = os.getenv("MODEL")

        if not self.openai_api_key:
            raise ValueError(" 未找到 OpenAI API Key, 请在 .env 文件中设置 OPENAI_API_KEY")

        self.client = OpenAI(api_key=self.openai_api_key, base_url=self.base_url)


    async def process_query(self, query: str) -> str:
        """调用 OpenAI API 处理用户查询"""
        messages = [
                {"role": "system", "content": "你是智能助手，帮助用户回答问题。"},
                {"role": "user", "content": query}
                ]

        try:
            #调用 OpenAI API
            response = await asyncio.get_event_loop().run_in_executor(
                    None,
                    lambda: self.client.chat.completions.create(
                        model=self.model,
                        messages=messages,
                        tools=self.tools
                        )
                    )

            #处理返回的内容
            content = response.choices[0]
            if content.finish_reason == "tool_calls":
                # 如何是需要使用工具，就解析工具

                tool_call = content.message.tool_calls[0]
                tool_name = tool_call.function.name
                tool_args = json.loads(tool_call.function.arguments)

                # 执行工具
                async with streamablehttp_client(url=self.mcp_server_url) as (read,write,get_session_id):
                    async with ClientSession(read,write) as session:
                        # 初始化会话(必须步骤)
                        await session.initialize()
                        result = await session.call_tool(tool_name, tool_args)
                        print(f"\n\n[Calling tool {tool_name} with args {tool_args}\n\n")

                # 将模型返回的调用哪个工具数据和工具执行完后的数据都存入messages 中
                messages.append(content.message.model_dump())
                messages.append({
                    "role": "tool",
                    "content": result.content[0].text,
                    "tool_call_id": tool_call.id,
                    })

                # 将上面的结果再返回给大模型用于生成最终结果
                response = self.client.chat.completions.create(
                        model=self.model,
                        messages=messages
                        )
                return response.choices[0].message.content

            return response.choices[0].message.content
        except Exception as e:
            return f" 调用 OpenAI API 时出错: {str(e)}"


    async def initialize(self):
        async with streamablehttp_client(url=self.mcp_server_url) as (read,write,get_session_id):
          async with ClientSession(read,write) as session:
            # 初始化会话(必须步骤)
            await session.initialize()
            # 获取工具列表
            response = await session.list_tools()
            self.tools = [self._convert_tool(t) for t in response.tools]

    def _convert_tool(self,tool):
        # 将MCP 工具转为兼容格式
        return {
                "type": "function",
                "function": {
                    "name": tool.name,
                    "description": tool.description,
                    "parameters": tool.inputSchema
                }
            }


    async def chat_loop(self):
        """ 运行交互聊天循环 """
        print("\n MCP 客户端已启动！输入 'quit' 退出")

        while True:
            try:
                query = input("\nQuery: ").strip()
                if query.lower() == 'quit':
                    break

                response = await self.process_query(query)

                print(f"\n OpenAI: {response}")

            except Exception as e:
                print(f"\n 发生错误：{str(e)}")


    async def cleanup(self):
        """ 清理资源 """
        await self.exit_stack.aclose()
        #if self.session:
        #    await self.session.close()

async def main():
    client = MCPClient()

    try:
        #await client.connect_to_server()
        await client.initialize()
        await client.chat_loop()
    finally:
        await client.cleanup()

if __name__ == "__main__":
    asyncio.run(main())
