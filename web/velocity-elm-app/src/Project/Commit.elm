module Project.Commit exposing (Commit, connectionSelectionSet, hash, selectionSet)

import Api.Compiled.Object
import Api.Compiled.Object.Commit as Commit
import Api.Compiled.Object.CommitConnection as CommitConnection
import Api.Compiled.Object.CommitEdge as CommitEdge
import Connection exposing (Connection)
import Edge exposing (Edge)
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)
import PageInfo exposing (PageInfo)
import Project.Commit.Hash as Hash exposing (Hash)


type Commit
    = Commit Internals


type alias Internals =
    { hash : Hash

    --    , author : String
    --    , date : Time.Posix
    , message : String
    }



-- SERIALIZATION


connectionSelectionSet : SelectionSet (Connection Commit) Api.Compiled.Object.CommitConnection
connectionSelectionSet =
    SelectionSet.map2 Connection
        (CommitConnection.pageInfo PageInfo.selectionSet)
        (CommitConnection.edges edgeSelectionSet
            |> SelectionSet.nonNullOrFail
            |> SelectionSet.nonNullElementsOrFail
        )


edgeSelectionSet : SelectionSet (Edge Commit) Api.Compiled.Object.CommitEdge
edgeSelectionSet =
    SelectionSet.succeed Edge.fromSelectionSet
        |> with CommitEdge.cursor
        |> with (SelectionSet.nonNullOrFail <| CommitEdge.node selectionSet)


selectionSet : SelectionSet Commit Api.Compiled.Object.Commit
selectionSet =
    SelectionSet.map Commit internalSelectionSet


internalSelectionSet : SelectionSet Internals Api.Compiled.Object.Commit
internalSelectionSet =
    SelectionSet.succeed Internals
        |> SelectionSet.with Hash.selectionSet
        |> SelectionSet.with Commit.message



-- INFO


hash : Commit -> Hash
hash (Commit commit) =
    commit.hash
