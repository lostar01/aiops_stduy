import asyncio
from mcp import ClientSession
from mcp.client.sse import sse_client

async def main():
    async with sse_client("http://localhost:30002/sse") as (read_stream,write_stream):
        async with ClientSession(read_stream, write_stream) as session:
            await session.initialize()


            # 列出可用工具
            tools = await session.list_tools()
            print("tools: ",tools)

            # 调用时间工具
            time_result = await session.call_tool(
                    "get_server_time",
                    {"format": "%Y-%m-%d %H:%M:%S"}
                    )
            print("服务时间:", time_result.content[0].text)

            # 调用网页抓取工具
            web_result = await session.call_tool(
                    "fetch_webpage",
                    {"url": "https://docs.aws.amazon.com/redshift/latest/mgmt/serverless-billing-reserved.html"}
                    )

            print("网页内容摘要:", web_result.content[0].text)

asyncio.run(main())

