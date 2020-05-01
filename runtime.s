# runtime
.text

# Start of program
.global _start
_start:
  callq runtime.heapInit
  callq main.main
  # exit(0)
  movq $0, %rdi # arg1
  movq $60, %rax
  syscall
# End of program

runtime.printstring:
  movq  8(%rsp), %rdi # arg0:ptr
  movq 16(%rsp), %rsi # arg1:len
  pushq %rsi
  pushq %rdi
  pushq $2 # stderr
  pushq $1 # sys_write
  callq syscall.Syscall
  addq $8 * 4, %rsp
  ret

syscall.Write:
  movq  8(%rsp), %rdx # arg0:fd
  movq 16(%rsp), %rsi # arg1:ptr
  movq 24(%rsp), %rdi # arg2:len
  pushq %rdi
  pushq %rsi
  pushq %rdx
  pushq $1  # sys_write
  callq syscall.Syscall
  addq $8 * 4, %rsp
  ret

syscall.Syscall:
  movq   8(%rsp), %rax
  movq  16(%rsp), %rdi
  movq  24(%rsp), %rsi
  movq  32(%rsp), %rdx
  syscall
  ret
