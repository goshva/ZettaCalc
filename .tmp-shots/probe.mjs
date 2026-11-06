import { createRequire } from 'node:module'
const require = createRequire('C:/Temp/shotter/')
const puppeteer = require('puppeteer-core')
const browser = await puppeteer.launch({
  executablePath: 'C:/Program Files/Google/Chrome/Application/chrome.exe',
  headless: true,
  pipe: true,
  userDataDir: `C:/Temp/pptr-${Date.now()}`,
  args: ['--no-sandbox', '--disable-gpu'],
})
const page = await browser.newPage()
await page.goto('http://localhost:5173/flow?step=0', { waitUntil: 'domcontentloaded' })
console.log('title:', await page.title())
console.log('wrap:', await page.evaluate(() => document.querySelector('.flow-page .flow-wrap')?.getBoundingClientRect().width))
await browser.close()
