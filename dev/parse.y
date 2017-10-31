%{

package main

import (
    "github.com/eddiejessup/gnex/lex"
)

%}

// Fields in this union become the fields of a structure
// '${PREFIX}SymType', a reference to which is passed to the lexer.
%union{
    val string
    result Result
    valCharCat lex.CharCat
    valCall lex.ControlSequenceCall
}

// Any non-terminal which returns a value needs a type, which is
// really a field name in the above union struct.
%type <result> result

// The same applies to terminals.
%token <valCharCat> CHAR_CAT
%token <valCall> CONTROL_SEQUENCE PAR END

%%

command : result
        {
            yylex.(*YaLexer).result = $1
        }
    ;
result : CHAR_CAT
        {
            $$ = $1
        }
    |   PAR
        {
            $$ = $1
        }
    |   END
        {
            $$ = $1
        }
    ;
%%
