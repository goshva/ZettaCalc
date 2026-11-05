/**
 * Закрытие <dialog> по клику на затемнение (как в оригинале: сравнение координат
 * с getBoundingClientRect). Вешается на @click диалога.
 */
export function closeOnBackdrop(e: MouseEvent) {
  const box = e.currentTarget as HTMLDialogElement | null
  if (!box || e.target !== box) return
  const r = box.getBoundingClientRect()
  if (e.clientX < r.left || e.clientX > r.right || e.clientY < r.top || e.clientY > r.bottom) {
    box.close()
  }
}
