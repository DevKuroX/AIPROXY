(function(){
  try{
    var t=localStorage.getItem('theme');
    if(!t&&window.matchMedia('(prefers-color-scheme:dark)').matches)t='dark';
    if(t==='dark')document.documentElement.classList.add('dark')
  }catch(e){}
})()
