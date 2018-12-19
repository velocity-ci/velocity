module Build.Id exposing (Id, decoder, toString, urlParser)

import Json.Decode as Decode exposing (Decoder)
import Url.Parser exposing (Parser)



-- TYPES


type Id
    = Id String



-- CREATE


urlParser : Parser (Id -> a) a
urlParser =
    Url.Parser.custom "ID" (\str -> Just (Id str))


decoder : Decoder Id
decoder =
    Decode.map Id Decode.string



-- TRANSFORM


toString : Id -> String
toString (Id str) =
    str
