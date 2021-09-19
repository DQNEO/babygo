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

# see https://man7.org/linux/man-pages/man2/clone.2.html
#  long clone(
#       unsigned long flags,
#       void *stack,
#       int *parent_tid,
#       int *child_tid,
#       unsigned long tls);

// func int clone(int flags, uintptr stack, uintptr fn)
runtime.clone:
  movq 24(%rsp), %r12
  movl $56, %eax # sys_clone
  movq 8(%rsp), %rdi # flags
  movq $0, %rsi # stack
  movq $0, %rdx # ptid
  movq $0, %r10 # chtid
  movq $0, %r8 # tls
  syscall

  cmpq $0, %rax # rax is pid for parent, 0 for child
  je .child # jump if child

  RET # return if parent

.child:
  callq *%r12
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

