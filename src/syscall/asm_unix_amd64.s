.text

// func Syscall(trap, a1, a2, a3 uintptr) uintptr
.global syscall.Syscall
syscall.Syscall:
  movq   8(%rsp), %rax # trap (syscall number)
  movq  16(%rsp), %rdi # a1
  movq  24(%rsp), %rsi # a2
  movq  32(%rsp), %rdx # a3
  syscall
  movq %rax, 40(%rsp) # r0 uintptr
  ret
