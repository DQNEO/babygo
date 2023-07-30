.text

// func Syscall(trap, a1, a2, a3 uintptr) uintptr
.global syscall.Syscall
syscall.Syscall:
  movq   8(%rsp), %rax # syscall number
  movq  16(%rsp), %rdi # arg0
  movq  24(%rsp), %rsi # arg1
  movq  32(%rsp), %rdx # arg2
  syscall
  movq %rax, 40(%rsp) # r0 uintptr
  ret
