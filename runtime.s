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
  movq $2,       %rdi #             arg0:fd
  movq  8(%rsp), %rsi # arg0:ptr -> arg1:buf
  movq 16(%rsp), %rdx # arg1:len -> arg2:len
  movq $1, %rax # sys_write
  syscall
  ret

syscall.Write:
  movq  8(%rsp), %rdi # arg0:number -> arg0:fd
  movq 16(%rsp), %rsi # arg1:ptr    -> arg1:buf
  movq 24(%rsp), %rdx # arg2:len    -> arg2:len
  movq $1, %rax # sys_write
  syscall
  ret
