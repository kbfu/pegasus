from multiprocessing import Pool
import aiohttp
import uvloop
import asyncio
from ept.core import date_util
from ept.db.log import Log


async def async_send():
    start_time = date_util.timestamp_now()
    async with aiohttp.ClientSession() as session:
        try:
            async with session.get('http://127.0.0.1:60001/get') as resp:
                resp_text = await resp.text()
                end_time = date_util.timestamp_now()
                elapsed_time = end_time - start_time
                Log().add('test', resp_text, resp.headers, elapsed_time,
                          start_time, end_time, resp.status)
        except Exception as e:
            end_time = date_util.timestamp_now()
            elapsed_time = end_time - start_time
            Log().add('test', resp_text, resp.headers, elapsed_time,
                      start_time, end_time, resp.status, e)


def run(n):
    asyncio.set_event_loop_policy(uvloop.EventLoopPolicy())
    loop = asyncio.get_event_loop()
    loop.run_until_complete(asyncio.gather(
        *[async_send() for _ in range(n)]
    ))
    loop.close()


if __name__ == '__main__':
    with Pool(processes=2) as pool:
        results = [pool.apply_async(run, (16, )) for _ in range(250)]
        for r in results:
            r.wait()