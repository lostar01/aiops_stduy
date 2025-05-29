import asyncio
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

# 配置服务器参数
server_params = StdioServerParameters(
        command = "python3",
        args = ["mcp-stdio-server.py"],
        env = {"PYTHONPATH": "."}  #添加环境变量
        )


async def main():
    async with stdio_client(server_params) as (read,write):
      async with ClientSession(read, write) as session:
        await session.initialize()  #初始化会话

        # 查看可用工具
        tools = await session.list_tools()
        print(f"可用工具: {[tool.name for tool in tools.tools]}")


        # 调用加法工具
        add_result = await session.call_tool("add", {"a": 4, "b": 6})
        print(f"4 +6 = { add_result.content[0].text }")

        # 调用乘法工具
        multiply_result = await session.call_tool("multiply", {"a": 11, "b": 14})
        print(f"11 * 14 = {multiply_result.content[0].text}")



if __name__ == "__main__":
    asyncio.run(main())

