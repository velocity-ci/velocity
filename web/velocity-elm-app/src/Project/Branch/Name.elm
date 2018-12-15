module Project.Branch.Name exposing (Name, decoder, default, name, text, toString, urlParser)

import Element exposing (..)
import Json.Decode as Decode exposing (Decoder)
import Url.Parser exposing (Parser)



-- TYPES


type Name
    = Name String


name : String -> Name
name str =
    Name str


default : Name
default =
    Name "master"



-- INFO


toString : Name -> String
toString (Name str) =
    str


text : Name -> Element msg
text (Name str) =
    Element.text str



-- CREATE


urlParser : Parser (Name -> a) a
urlParser =
    Url.Parser.custom "Name" (\str -> Just (Name str))


decoder : Decoder Name
decoder =
    Decode.map Name Decode.string
