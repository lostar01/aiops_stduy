from mcp.server.fastmcp import FastMCP
from pydantic import Field
import logging

# 配置日志
logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")
logger = logging.getLogger("MathServer")

# 创建 FastMCP 实例
mcp = FastMCP("MatServer")

@mcp.tool(name="add", description="加法运算")
def add(
        a: int = Field(description="被加数"),
        b: int = Field(description="加数")
        ) -> int:
    """ 返回两个整数的和 """
    logger.info(f"执行加法: {a} + {b}")
    return a+b

@mcp.tool(name="multiply", description="乘法运算")
def add(
        a: int = Field(description="被乘数"),
        b: int = Field(description="乘数")
        ) -> int:
    """ 返回两个整数的乘积 """
    logger.info(f"执行乘法: {a} x {b}")
    return a*b

if __name__ == "__main__":
    mcp.run(transport="stdio")  # 启动STDIO 模式
