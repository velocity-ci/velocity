module Project.Commit.Hash exposing (Hash, decoder, toString, urlParser)

import Json.Decode as Decode exposing (Decoder)
import Url.Parser exposing (Parser)


-- TYPES


type Hash
    = Hash String



-- INFO


toString : Hash -> String
toString (Hash str) =
    str



-- CREATE


urlParser : Parser (Hash -> a) a
urlParser =
    Url.Parser.custom "SLUG" (\str -> Just (Hash str))


decoder : Decoder Hash
decoder =
    Decode.map Hash Decode.string
