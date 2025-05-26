from mcp.client.streamable_http import streamablehttp_client
from mcp import ClientSession
import asyncio

async def main():
    server_url = "http://127.0.0.1:30001/mcp"

    async with streamablehttp_client(url=server_url) as (read,write,get_session_id):
      async with ClientSession(read,write) as session:
        # 初始化会话(必须步骤)
        await session.initialize()

        # 获取工具列表
        tools = await session.list_tools()
        print(f"可用工具: {[t for t  in tools]}")

        ## 调用流式工具
        #async for chunk in session.call_tool_stream(
        #            "text_streamer",
        #            {"text": "HelloStreamableHTTP"}
        #        ):

        #            print(f"实时响应: {chunk.decode().strip()}")

        # 获取最终结果
        final_result = await session.call_tool(
                #"text_streamer",
                "add",
                {"a": 1 , "b": 2}
                )
        print(f"最终结果: {final_result.content}")



asyncio.run(main())

