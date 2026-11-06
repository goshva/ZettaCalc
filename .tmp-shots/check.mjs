// Проверка вёрстки flow-страницы на широких экранах.
// Запуск: node .tmp-shots/check.mjs [--shots]
// puppeteer-core подключён из временной папки (в репозиторий не ставим).
import { createRequire } from 'node:module'
import { mkdirSync } from 'node:fs'

const require = createRequire('C:/Temp/shotter/')
const puppeteer = require('puppeteer-core')

const CHROME = 'C:/Program Files/Google/Chrome/Application/chrome.exe'
const OUT = 'C:/soft/ZettaCalc/.tmp-shots'
const withShots = process.argv.includes('--shots')

const TARGETS = [
  { name: 'A', base: 'http://localhost:5173/flow', step: (n) => `?step=${n}` },
  { name: 'B', base: 'http://localhost:5174/#/flow', step: (n) => `?step=${n}` },
]
const WIDTHS = [1440, 1920, 2560]
const STEPS = [0, 1, 2, 3, 4]

mkdirSync(OUT, { recursive: true })

const browser = await puppeteer.launch({
  executablePath: CHROME,
  headless: true,
  userDataDir: `C:/Temp/pptr-${Date.now()}`,
  args: ['--no-sandbox', '--hide-scrollbars', '--disable-gpu'],
})

const problems = []

const collect = () => {
  const vw = window.innerWidth
  const wrap = document.querySelector('.flow-page .flow-wrap')
  const cards = [...document.querySelectorAll('.flow-page .flow-card')]
  const overflowing = []
  for (const el of document.querySelectorAll('.flow-page *')) {
    const r = el.getBoundingClientRect()
    if (r.width === 0 && r.height === 0) continue
    if (r.right > vw + 1 || r.left < -1) {
      overflowing.push({
        cls: (el.className || '').toString().slice(0, 60),
        tag: el.tagName,
        left: Math.round(r.left),
        right: Math.round(r.right),
      })
    }
  }
  return {
    vw,
    docScrollWidth: document.documentElement.scrollWidth,
    wrapWidth: wrap ? Math.round(wrap.getBoundingClientRect().width) : null,
    cardWidth: cards.length ? Math.round(cards[0].getBoundingClientRect().width) : null,
    cardCount: cards.length,
    overflowCount: overflowing.length,
    overflowing: overflowing.slice(0, 6),
    active: document.querySelector('.flow-page .flow-step.active')?.textContent?.trim(),
    h2: document.querySelector('.flow-page .flow-card h2')?.textContent?.trim(),
  }
}

for (const t of TARGETS) {
  for (const width of WIDTHS) {
    for (const step of STEPS) {
      const page = await browser.newPage()
      await page.setViewport({ width, height: 1000, deviceScaleFactor: 1 })
      await page.goto(t.base + t.step(step), { waitUntil: 'networkidle2', timeout: 30000 })
      await page.waitForSelector('.flow-page', { timeout: 10000 })
      await new Promise((r) => setTimeout(r, 400))

      const data = await page.evaluate(collect)

      let auction = null
      if (step === 4) {
        const clicked = await page.evaluate(() => {
          const btn = [...document.querySelectorAll('.flow-page button')].find((b) =>
            b.textContent.includes('Заказать ЦПР и начать аукцион'),
          )
          if (!btn) return false
          btn.click()
          return true
        })
        if (clicked) {
          await new Promise((r) => setTimeout(r, 400))
          auction = await page.evaluate(() => {
            const bids = [...document.querySelectorAll('.flow-page .bid-card')]
            const res = collect()
            return {
              bidCount: bids.length,
              bidWidths: bids.map((b) => Math.round(b.getBoundingClientRect().width)),
              overflowCount: res.overflowCount,
              overflowing: res.overflowing,
              cardWidth: res.cardWidth,
            }
          })
        }
      }

      const label = `${t.name} ${width} step=${step}`
      const notes = []
      if (data.docScrollWidth > data.vw + 1) notes.push(`гориз. скролл doc=${data.docScrollWidth} > vw=${data.vw}`)
      if (data.overflowCount) notes.push(`за краем ${data.overflowCount}: ${JSON.stringify(data.overflowing)}`)
      if (data.cardWidth && data.cardWidth > 1360) notes.push(`карточка шире 1360: ${data.cardWidth}`)
      if (data.cardCount === 0) notes.push('нет карточек')
      if (auction) {
        if (auction.bidCount !== 3) notes.push(`ставок ${auction.bidCount}`)
        if (auction.overflowCount) notes.push(`аукцион за краем ${auction.overflowCount}: ${JSON.stringify(auction.overflowing)}`)
      }

      const line = `${label}: wrap=${data.wrapWidth} card=${data.cardWidth} шаг="${data.active}" h2="${data.h2}"${auction ? ` bids=${auction.bidCount}${JSON.stringify(auction.bidWidths)}` : ''}`
      if (notes.length) {
        problems.push(`${label}: ${notes.join(' | ')}`)
        console.log('ПРОБЛЕМА ' + line)
        console.log('         ' + notes.join(' | '))
      } else {
        console.log('ok  ' + line)
      }

      if (withShots) {
        await page.screenshot({ path: `${OUT}/${t.name}-${width}-step${step}${auction ? '-auction' : ''}.png` })
      }
      await page.close()
    }
  }
}

await browser.close()
console.log('\nИтого проблем: ' + problems.length)
for (const p of problems) console.log(' - ' + p)
