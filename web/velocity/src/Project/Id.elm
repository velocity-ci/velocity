module Project.Id
    exposing
        ( Dict
        , Id
        , decoder
        , empty
        , get
        , insert
        , routePieces
        , urlParser
        , selectionSet
        )

import Dict as BaseDict
import Json.Decode as Decode exposing (Decoder)
import Url.Parser exposing (Parser)
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, hardcoded, with)
import Api.Compiled.Object
import Api.Compiled.Object.Project as Project
import Api.Compiled.Scalar as Scalar


-- TYPES


type Id
    = Id String


type Dict e
    = Dict (BaseDict.Dict String e)



-- CREATE


urlParser : Parser (Id -> a) a
urlParser =
    Url.Parser.custom "ID" (\str -> Just (Id str))


decoder : Decoder Id
decoder =
    Decode.map Id Decode.string


selectionSet : SelectionSet Id Api.Compiled.Object.Project
selectionSet =
    SelectionSet.map (\(Scalar.Id id) -> Id id) Project.id



-- ROUTE


routePieces : Id -> List String
routePieces (Id str) =
    [ "project", str ]



-- DICT


empty : Dict e
empty =
    Dict BaseDict.empty


insert : Id -> e -> Dict e -> Dict e
insert (Id str) e (Dict d) =
    BaseDict.insert str e d
        |> Dict


get : Id -> Dict e -> Maybe e
get (Id str) (Dict d) =
    BaseDict.get str d
