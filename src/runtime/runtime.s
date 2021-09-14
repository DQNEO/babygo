# runtime.s
.text

# (runtime/asm_amd64.s)
_rt0_amd64:
  movq 0(%rsp), %rdi # argc
  leaq 8(%rsp), %rsi # argv
  jmp runtime.rt0_go


// func Write(fd int, p []byte) int
.global runtime.Write
runtime.Write:
  movq  8(%rsp), %rax # arg0:fd
  movq 16(%rsp), %rdi # arg1:ptr
  movq 24(%rsp), %rsi # arg2:len
  subq $8, %rsp
  pushq %rsi
  pushq %rdi
  pushq %rax
  pushq $1  # sys_write
  callq runtime.Syscall
  addq $8 * 4, %rsp # reset args area
  popq %rax # retval
  movq %rax, 32(%rsp) # r0 int
  ret

.global runtime.printstring
runtime.printstring:
  movq  8(%rsp), %rdi # arg0:ptr
  movq 16(%rsp), %rsi # arg1:len
  subq $8, %rsp
  pushq %rsi
  pushq %rdi
  pushq $2 # stderr
  pushq $1 # sys_write
  callq runtime.Syscall
  addq $8 * 4, %rsp
  popq %rax # retval
  ret

// func Syscall(trap, a1, a2, a3 uintptr) uintptr
.global runtime.Syscall
runtime.Syscall:
  movq   8(%rsp), %rax # syscall number
  movq  16(%rsp), %rdi # arg0
  movq  24(%rsp), %rsi # arg1
  movq  32(%rsp), %rdx # arg2
  syscall
  movq %rax, 40(%rsp) # r0 uintptr
  ret

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

