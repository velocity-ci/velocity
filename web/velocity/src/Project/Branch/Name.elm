module Project.Branch.Name exposing (Name, default, name, queryParser, selectionSet, text, toString, urlParser)

import Api.Compiled.Object
import Api.Compiled.Object.Branch as Branch
import Element exposing (..)
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)
import Url.Parser exposing (Parser)
import Url.Parser.Query as QueryParser



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


queryParser : QueryParser.Parser (Maybe String) -> QueryParser.Parser (Maybe Name)
queryParser =
    QueryParser.map (Maybe.map Name)



-- SERIALIZATION --


selectionSet : SelectionSet Name Api.Compiled.Object.Branch
selectionSet =
    SelectionSet.succeed Name
        |> with Branch.name
