module Project.Commit.Hash exposing (Hash, decoder, toString, urlParser, selectionSet)

import Api.Compiled.Object
import Api.Compiled.Object.Commit as Commit
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)

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

selectionSet : SelectionSet Hash Api.Compiled.Object.Commit
selectionSet =
    SelectionSet.map Hash Commit.sha


urlParser : Parser (Hash -> a) a
urlParser =
    Url.Parser.custom "SLUG" (\str -> Just (Hash str))


decoder : Decoder Hash
decoder =
    Decode.map Hash Decode.string
