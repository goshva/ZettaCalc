;
favs=function(){wrap51(crumb('Избранное')+'<h1>Избранные товары</h1><div class="grid section">'+ALL51.filter(p=>favorites.includes(p.id)).map(genericCard51).join('')+'</div>')}