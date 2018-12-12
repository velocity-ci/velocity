module Project.Slug exposing
    ( Dict
    , Slug
    , decoder
    , empty
    , get
    , insert
    , routePieces
    , toString
    , urlParser
    )

import Dict as BaseDict
import Json.Decode as Decode exposing (Decoder)
import Url.Parser exposing (Parser)



-- TYPES


type Slug
    = Slug String


type Dict e
    = Dict (BaseDict.Dict String e)



-- CREATE


urlParser : Parser (Slug -> a) a
urlParser =
    Url.Parser.custom "SLUG" (\str -> Just (Slug str))


decoder : Decoder Slug
decoder =
    Decode.map Slug Decode.string



-- TRANSFORM


toString : Slug -> String
toString (Slug str) =
    str



-- ROUTE


routePieces : Slug -> List String
routePieces (Slug str) =
    [ "project", str ]



-- DICT


empty : Dict e
empty =
    Dict BaseDict.empty


insert : Slug -> e -> Dict e -> Dict e
insert (Slug str) e (Dict d) =
    BaseDict.insert str e d
        |> Dict


get : Slug -> Dict e -> Maybe e
get (Slug str) (Dict d) =
    BaseDict.get str d
