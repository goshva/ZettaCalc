import { onBeforeUnmount, onBeforeMount } from 'vue'

/**
 * Подключает CSS отдельного приложения прототипа А на время жизни страницы.
 *
 * В оригинале таблица стилей приложения подменяла таблицу главной целиком
 * (`mount52`: `style52.textContent = info.css`), а постоянной оставалась
 * только оболочка (`/api/css/a/shell`). Здесь то же поведение: пока открыта
 * страница приложения, CSS главной выключен — иначе `.gallery` из главной
 * навязывает интерьерам сетку в 4 колонки, а `.close` ломает галерею.
 */

// Сколько страниц приложений сейчас на экране. Счётчик, а не флаг: при
// переходе между двумя приложениями старый компонент размонтируется уже
// после монтирования нового, и без счётчика он включил бы CSS главной.
const mountedApps = new Set<string>()

function syncMainStyles() {
  document.querySelectorAll<HTMLLinkElement>('link[data-app-css="main"]').forEach((link) => {
    link.disabled = mountedApps.size > 0
  })
}

export function useAppStyles(app: 'interiors' | 'center', htmlClass: string) {
  const own = `link[data-app-css="${app}"]`

  onBeforeMount(() => {
    document.documentElement.classList.add(htmlClass)
    mountedApps.add(app)
    syncMainStyles()
    if (document.querySelector(own)) return
    const link = document.createElement('link')
    link.rel = 'stylesheet'
    link.href = `/api/apps/${app}/css`
    link.dataset.appCss = app
    document.head.appendChild(link)
  })

  onBeforeUnmount(() => {
    document.documentElement.classList.remove(htmlClass)
    mountedApps.delete(app)
    syncMainStyles()
    document.querySelector(own)?.remove()
  })
}
