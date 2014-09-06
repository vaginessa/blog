Title: Case study: optimizing disassembler in v8

Optimizing v8
=============

This case study shows that even mature code can have opportunities for
optimization. When I noticed an opportunity to optimize disassembler in
v8, it was already at version 1.3.6 and had more than 2700 checkins.

Sources of waste
----------------

Disassembler is driven by 2 data structures: InstructionDesc and
InstructionTable.

@includesrc disasm-ia32.cc 139 143

@includesrc disasm-ia32.cc 47 51

At runtime, InstructionDesc structure describes instruction for a given
byte. InstructionTable class contains a 256 element array of
InstructionDesc structures. On 32-bit architecture
sizeof(InstructionDesc) is 12 bytes so the whole array consumes 12\*256
= 3072 bytes.

This table is built at runtime from several arrays of ByteMnemonic
struct. There are 41 elements in total and since sizeof(ByteMnemonic) is
12, they use 492 bytes of static space.

Total space used: 3564 bytes.

Eliminating the waste
---------------------

### Removing unnecesary code and data

First observation is that InstructionDesc array could just as well be
embedded as static data. We could get rid of 492 bytes of ByteMnemonic
data and the code that builds InstructionDesc array out of ByteMnemonic
data.

Presumably it was done that way for readability. We certainly wouldn’t
want to create the array data manually - it would be tedious and the
data would be inscrutable. We can fix that with a python script that
generates C code statically defining InstructionDesc array using the
same ByteMnemonic data currently part of C code. The resulting data will
still be inscrutable but the process of generating it will be as
readable as before.

### Singleton class is just a bunch of scoped functions and data

Another observation is that InstructionTable class is unnecessary.
There’s only one instance of it and when we switched to having
InstructionDesc data defined at compilation time, it gets reduced to
only one function. We can replace InstructionTable class with the one
remaining function.

### Using the most efficient encoding

Another observation is that encoding of data is inefficient.
InstructionType enum has 8 possible values and could be encoded in 3
bits, but an enum requires 4 bytes. OperandOrder enum has 3 possible
values and coudl be encoded in 2 bits, but uses 4 bytes. Using bitfield
we could replace 8 bytes with just one:

<code class="cpp"><pre>\
struct InstructionDesc {\
 const char\* mnem;\
 unsigned char type : 3;\
 unsigned char op\_order\_ : 2;\
};

</pre>
</code>

### Split arrays to avoid alignment padding

Unfortunately even though mnem only uses 4 bytes and type + op\_order\_
use 1 byte, sizeof(InstructionDesc) is 8 bytes, not 5. On 64-bit
processor it would be even worse and the size would be 16. Elements in a
C struct are aligned to their natural size so e.g. pointers are aligned
to 4 or 8 bytes (on 32-bit and 64-bit respectively). The size of the
whole struct also needs to be rounded to the alignement of the first
element to allow making
