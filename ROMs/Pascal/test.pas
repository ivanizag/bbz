program charSet(output);
{ Prints the ASCII character set }
var
   ch : char;

begin
   for ch:=chr(32) to chr(126) do
      write(ch);
    writeln
end.
