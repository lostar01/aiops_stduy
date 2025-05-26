#!/usr/bin/env python3

from mcp.server.fastmcp import FastMCP, Context
import asyncio

# Create an MCP server
mcp = FastMCP("Demo", host="0.0.0.0", port=30001)

# Add an addition tool
@mcp.tool(name="add")
def add(a: int, b: int) -> int:
    """ Add two numbers """
    return a + b


@mcp.tool(name="get_weather")
def get_weather(city: str) -> dict:
    """ 获取天气城市的当天天气数据
    Args:
       city: 城市中文名称（如 "北京")
    Returns:
       包含温度、天气状况的字典
    """
    return {
      "temperature": 24.5,
      "conditon": "晴",
      "humidity": 65
    }

@mcp.tool(name="text_streamer")
async def stream_tool(ctx: Context, text: str):
    """模拟流式生成"""
    #发送初始化进度
    await ctx.report_progress(0.0,5.0,"开始处理")

    #分阶段生成文本
    for i in range(1,6):
        await asyncio.sleep(1)
        #发送中间结果(流式)
        yield f"data: 第{i}次生成: {text[:i*2]}\n\n"
        #更新进度条
        await ctx.report_progress(float(i),5.0, f"完成阶段{i}/5")

# Add a dynamic greeting resource
@mcp.resource("greeting://{name}")
def get_greeting(name: str) -> str:
    """ Get a personalized greeting """
    return f"Hello, {name}!"

if __name__ == '__main__':
    mcp.run(transport="streamable-http")

