function copyCode() {
  const code = document.querySelector('pre code').innerText;
  navigator.clipboard.writeText(code);
}
