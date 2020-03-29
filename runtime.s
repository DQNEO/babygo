// runtime
.text
.global _start
_start:
  movq $42, %rdi
  movq $60, %rax
  syscall
