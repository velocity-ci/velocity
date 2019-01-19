module Project.Branch exposing (Branch, text, selectionSet)

import Element exposing (..)
import Project.Branch.Name as Name exposing (Name)
import Api.Compiled.Object
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)


type Branch
    = Branch Internals


type alias Internals =
    { name : Name
    }



-- INFO


text : Branch -> Element msg
text (Branch { name }) =
    Name.text name



-- SERIALIZATION --


selectionSet : SelectionSet Branch Api.Compiled.Object.Branch
selectionSet =
    SelectionSet.succeed Branch
        |> with internalSelectionSet


internalSelectionSet : SelectionSet Internals Api.Compiled.Object.Branch
internalSelectionSet =
    SelectionSet.succeed Internals
        |> with Name.selectionSet
