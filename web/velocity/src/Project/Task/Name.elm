module Project.Task.Name exposing (Name, decoder, selectionSet, toString, urlParser)

import Api.Compiled.Object
import Api.Compiled.Object.Task as Task
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)
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


selectionSet : SelectionSet Name Api.Compiled.Object.Task
selectionSet =
    SelectionSet.map Name Task.name


urlParser : Parser (Name -> a) a
urlParser =
    Url.Parser.custom "Name" (\str -> Just (Name str))


decoder : Decoder Name
decoder =
    Decode.map Name Decode.string
