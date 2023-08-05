.text

// func Syscall(trap, a1, a2, a3 uintptr) (r1 uintptr, r2 uintptr, errno uintptr)
.global syscall.Syscall
syscall.Syscall:
  movq   8(%rsp), %rax # trap (syscall number)
  movq  16(%rsp), %rdi # a1
  movq  24(%rsp), %rsi # a2
  movq  32(%rsp), %rdx # a3
  syscall
  movq %rax, 40(%rsp) # r1 uintptr
  movq %rdx, 48(%rsp) # r2 uintptr
  movq %rax, 56(%rsp) # errno
  ret
