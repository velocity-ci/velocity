module Project.Task.Name exposing (Name, decoder, toString, urlParser)

import Json.Decode as Decode exposing (Decoder)
import Url.Parser exposing (Parser)


-- TYPES


type Name
    = Name String



-- INFO


toString : Name -> String
toString (Name str) =
    str



-- CREATE


urlParser : Parser (Name -> a) a
urlParser =
    Url.Parser.custom "Name" (\str -> Just (Name str))


decoder : Decoder Name
decoder =
    Decode.map Name Decode.string
