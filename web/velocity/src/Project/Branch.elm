module Project.Branch exposing (Branch, connectionSelectionSet, name, selectionSet, text)

import Api.Compiled.Object
import Api.Compiled.Object.BranchConnection as BranchConnection
import Api.Compiled.Object.BranchEdge as BranchEdge
import Connection exposing (Connection)
import Edge exposing (Edge)
import Element exposing (..)
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)
import PageInfo exposing (PageInfo)
import Project.Branch.Name as Name exposing (Name)


type Branch
    = Branch Internals


type alias Internals =
    { name : Name
    }



-- INFO


name : Branch -> Name
name (Branch internals) =
    internals.name


text : Branch -> Element msg
text (Branch internals) =
    Name.text internals.name



-- SERIALIZATION --


connectionSelectionSet : SelectionSet (Connection Branch) Api.Compiled.Object.BranchConnection
connectionSelectionSet =
    SelectionSet.map2 Connection
        (BranchConnection.pageInfo PageInfo.selectionSet)
        (BranchConnection.edges edgeSelectionSet
            |> SelectionSet.nonNullOrFail
            |> SelectionSet.nonNullElementsOrFail
        )


edgeSelectionSet : SelectionSet (Edge Branch) Api.Compiled.Object.BranchEdge
edgeSelectionSet =
    SelectionSet.succeed Edge.fromSelectionSet
        |> with BranchEdge.cursor
        |> with (SelectionSet.nonNullOrFail <| BranchEdge.node selectionSet)


selectionSet : SelectionSet Branch Api.Compiled.Object.Branch
selectionSet =
    SelectionSet.succeed Branch
        |> with internalSelectionSet


internalSelectionSet : SelectionSet Internals Api.Compiled.Object.Branch
internalSelectionSet =
    SelectionSet.succeed Internals
        |> with Name.selectionSet
