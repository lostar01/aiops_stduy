from fastapi import FastAPI, Request
from mcp.server.fastmcp import FastMCP
from mcp.server.sse import SseServerTransport
from starlette.routing import Mount
import uvicorn
import httpx

# 初始化MCP 服务
mcp = FastMCP(name="DemoServer")

@mcp.tool(name="get_server_time", description="获取服务器当前时间")
async def get_server_time(format: str = "%Y-%m-%d %H:%M:%S") -> str:
    """ 时间格式化工具 """
    from datetime import datetime
    return datetime.now().strftime(format)


@mcp.tool(name="fetch_webpage", description="获取网页内容")
async def fetch_webpage(url: str) -> str:
    """ 网页抓取工具 """
    async with httpx.AsyncClient() as client:
        response = await client.get(url)
        return response.text[:50000]  #返回前50000字符串


# 配置 FastAPI 应用
app = FastAPI()
sse_transport = SseServerTransport("/messages/")
app.router.routes.append(Mount("/messages", app=sse_transport.handle_post_message))

@app.get("/sse")
async def handle_sse_connection(request: Request):
    async with sse_transport.connect_sse(
            scope = request.scope,
            receive = request.receive,
            send = request._send
    ) as (read_stream,write_stream):

        await mcp._mcp_server.run(
                read_stream=read_stream,
                write_stream=write_stream,
                initialization_options=mcp._mcp_server.create_initialization_options()
                )


if __name__ == "__main__":
  uvicorn.run(app, host="0.0.0.0", port=30002)

