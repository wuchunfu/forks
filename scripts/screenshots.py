import asyncio
import os
from playwright.async_api import async_playwright

BASE = 'http://127.0.0.1:8080'
TOKEN = '12345678'
OUT = os.path.join(os.path.dirname(os.path.dirname(__file__)), 'images')

async def main():
    async with async_playwright() as p:
        browser = await p.chromium.launch()
        context = await browser.new_context(
            viewport={'width': 1440, 'height': 900},
            color_scheme='dark',
        )

        pg = await context.new_page()

        # 通过 ?token= 登录
        await pg.goto(f'{BASE}/?token={TOKEN}', wait_until='domcontentloaded', timeout=15000)
        await pg.wait_for_timeout(2000)
        await pg.evaluate('window.history.replaceState({}, "", "/")')

        # 首页
        print('截图: 首页')
        await pg.reload(wait_until='domcontentloaded', timeout=15000)
        await pg.wait_for_timeout(4000)
        await pg.add_style_tag(content='::-webkit-scrollbar { width: 0 !important; height: 0 !important; }')
        await pg.wait_for_timeout(300)
        await pg.screenshot(path=f'{OUT}/首页.png')
        print(f'  done ({os.path.getsize(f"{OUT}/首页.png")} bytes)')

        pages = [
            ('仓库列表', 'repos'),
            ('作者', 'authors'),
            ('趋势', 'trending'),
            ('活动', 'activities'),
            ('设置', 'settings'),
        ]

        for name, path in pages:
            print(f'截图: {name}')
            await pg.goto(f'{BASE}/{path}', wait_until='domcontentloaded', timeout=15000)
            await pg.wait_for_timeout(3000)
            await pg.add_style_tag(content='::-webkit-scrollbar { width: 0 !important; height: 0 !important; }')
            await pg.wait_for_timeout(300)
            await pg.screenshot(path=f'{OUT}/{name}.png')
            print(f'  done ({os.path.getsize(f"{OUT}/{name}.png")} bytes)')

        # 仓库详情：点击第一个仓库
        print('截图: 仓库详情')
        await pg.goto(f'{BASE}/repos', wait_until='domcontentloaded', timeout=15000)
        await pg.wait_for_timeout(2000)
        card = pg.locator('.repo-card').first
        try:
            await card.wait_for(state='visible', timeout=3000)
            await card.click()
            await pg.wait_for_timeout(2000)
        except:
            print('    (未找到仓库卡片)')
        await pg.screenshot(path=f'{OUT}/仓库详情.png')
        print(f'  done ({os.path.getsize(f"{OUT}/仓库详情.png")} bytes)')

        # 代码查看：从 API 获取一个已克隆仓库的 ID
        print('截图: 代码查看')
        resp = await pg.evaluate(f'''
            fetch('{BASE}/api/repos?status=cloned&page=1&page_size=1', {{
                headers: {{ 'Authorization': 'Bearer {TOKEN}' }}
            }}).then(r => r.json())
        ''')
        code_id = None
        try:
            code_id = resp['data']['list'][0]['id']
        except:
            print('    (未找到已克隆仓库，跳过)')
        if code_id:
            await pg.goto(f'{BASE}/code/{code_id}', wait_until='domcontentloaded', timeout=15000)
            await pg.wait_for_timeout(3000)
            # 点击第一个文件加载内容
            file_item = pg.locator('.file-tree .tree-item, .n-tree-node-content').first
            try:
                await file_item.wait_for(state='visible', timeout=3000)
                await file_item.click()
                await pg.wait_for_timeout(2000)
            except:
                pass
            await pg.add_style_tag(content='::-webkit-scrollbar { width: 0 !important; height: 0 !important; }')
            await pg.wait_for_timeout(300)
            await pg.screenshot(path=f'{OUT}/代码查看.png')
            print(f'  done ({os.path.getsize(f"{OUT}/代码查看.png")} bytes)')

        await browser.close()
        print('全部完成')

asyncio.run(main())
