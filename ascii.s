.data
msg:
   .ascii "arm64"
.text
.global main
main:
  sub sp, sp, 16
  mov x23, sp
  mov x9, 0
  mov x18, 0
  sub x23, x23, 8
  mov x7, 0x23b
  and x18, x7, 0xf
  lsr x7, x7, 4
  add x18, x18, 48
  add x9, x9, x18
  lsl x9, x9, 8
  and x18, x7, 0xf
  lsr x7, x7, 4
  add x18, x18, 48
  add x9, x9, x18

//  mov x9, 0x41
  str x9, [x23]
  mov x1, x23
  mov x2, 2
  mov x8, 64
  svc 0
  ret
