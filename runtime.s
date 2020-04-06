# runtime
.text

# Start of program
.global _start
_start:
  callq main.main
  # exit(0)
  movq $0, %rdi # arg1
  movq $60, %rax
  syscall
# End of program

runtime.printstring:
  movq $2, %rdi # stderr
  movq  8(%rsp), %rsi # arg0:buf -> arg1:buf
  movq 16(%rsp), %rdx # arg1:len -> arg0:len
  movq $1, %rax # sys_write
  syscall
  ret
