# runtime
.text
.global _start
_start:
  callq main.main

# .global os.Exit
os.Exit:
  movq 8(%rsp), %rdi # arg1
  movq $60, %rax
  syscall
# End of program
