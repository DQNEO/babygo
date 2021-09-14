.text

runtime.rt0_go:
  movq %rdi, %rax # argc
  movq %rsi, %rbx # argv

  subq $32, %rsp
  movq %rax, 16(%rsp) # argc
  movq %rbx, 24(%rsp) # argv

  movq 16(%rsp), %rax  # copy argc
  movq %rax, 0(%rsp)
  movq 24(%rsp), %rbx  # copy argv
  movq %rbx, 8(%rsp)
  callq runtime.args
  addq $32, %rsp

  movq %rdi, %rax # argc
  imulq $8,  %rax # argc * 8
  addq %rsp, %rax # stack top addr + (argc * 8)
  addq $16,  %rax # + 16 (skip null and go to next) => envp
  movq %rax, runtime.envp+0(%rip) # envp

  callq runtime.heapInit

  callq runtime.__initGlobals
  callq runtime.envInit

  callq main.__initGlobals

  callq os.init # set os.Args
  callq main.main

  movq $0, %rdi  # status 0
  movq $60, %rax # sys_exit
  syscall
  # End of program

runtime.args:
  movq  8(%rsp), %rax # argc
  movq 16(%rsp), %rbx # argv
  movq %rbx, runtime.__argv__+0(%rip)  # ptr
  movq %rax, runtime.__argv__+8(%rip)  # len
  movq %rax, runtime.__argv__+16(%rip) # cap
  ret

