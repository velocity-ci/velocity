module Project.Id exposing
    ( Dict
    , Id
    , decoder
    , empty
    , get
    , insert
    , routePieces
    , urlParser
    )

import Dict as BaseDict
import Json.Decode as Decode exposing (Decoder)
import Url.Parser exposing (Parser)



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
