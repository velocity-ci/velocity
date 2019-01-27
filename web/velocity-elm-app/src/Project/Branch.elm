module Project.Branch exposing (Branch, text, selectionSet)

import Element exposing (..)
import Project.Branch.Name as Name exposing (Name)
import Api.Compiled.Object
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)
import Api.Compiled.Object.BranchEdge as BranchEdge


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

type alias Connection a =
    { pageInfo : PageInfo
    , edges : List (Edge a)
    }

type alias PageInfo =
    { startCursor : String
    , endCursor : String
    , hasNextPage : Bool
    , hasPreviousPage : Bool
    }

type alias Edge a =
    { cursor : String
    , node : a
    }

edgeSelectionSet : SelectionSet (Edge (Maybe Branch)) Api.Compiled.Object.BranchEdge
edgeSelectionSet =
    SelectionSet.succeed Edge
        |> with BranchEdge.cursor
        |> with (BranchEdge.node selectionSet)

selectionSet : SelectionSet Branch Api.Compiled.Object.Branch
selectionSet =
    SelectionSet.succeed Branch
        |> with internalSelectionSet


internalSelectionSet : SelectionSet Internals Api.Compiled.Object.Branch
internalSelectionSet =
    SelectionSet.succeed Internals
        |> with Name.selectionSet
