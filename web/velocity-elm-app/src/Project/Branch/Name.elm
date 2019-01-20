module Project.Branch.Name exposing (Name, default, name, text, toString, urlParser, selectionSet)

import Element exposing (..)
import Url.Parser exposing (Parser)
import Api.Compiled.Object.Branch as Branch
import Api.Compiled.Object
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)


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



-- SERIALIZATION --


selectionSet : SelectionSet Name Api.Compiled.Object.Branch
selectionSet =
    SelectionSet.succeed Name
        |> with Branch.name
