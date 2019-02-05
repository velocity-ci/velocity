module Event exposing (Log, selectionSet)

import Api.Compiled.Object
import Api.Compiled.Object.Event as Event
import Api.Compiled.Object.EventConnection as EventConnection
import Api.Compiled.Object.EventEdge as EventEdge
import Connection exposing (Connection)
import Edge exposing (Edge)
import Element exposing (..)
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)
import KnownHost exposing (KnownHost)
import PageInfo exposing (PageInfo)
import Project.Id as ProjectId
import Set exposing (Set)
import Username exposing (Username)


type Log
    = Log Internals


type alias Internals =
    { seen : Set Id
    , events : Connection Event
    }


init : Connection Event -> Log
init connection =
    Log <|
        { seen = Set.empty
        , events = connection
        }


type Event
    = Event Username EventType


type EventType
    = KnownHostAdded KnownHost.Id
    | KnownHostVerified KnownHost.Id
    | ProjectAdded ProjectId.Id
    | Unknown


type Id
    = Id String


selectionSet : SelectionSet Log Api.Compiled.Object.EventConnection
selectionSet =
    SelectionSet.map Log internalSelectionSet


internalSelectionSet : SelectionSet Internals Api.Compiled.Object.EventConnection
internalSelectionSet =
    SelectionSet.map2 toInternals
        (EventConnection.pageInfo PageInfo.selectionSet)
        (EventConnection.edges edgeSelectionSet
            |> SelectionSet.nonNullOrFail
            |> SelectionSet.nonNullElementsOrFail
        )


toInternals : PageInfo -> List (Edge Event) -> Internals
toInternals pageInfo edges =
    { seen = Set.empty
    , events = Connection pageInfo edges
    }


edgeSelectionSet : SelectionSet (Edge Event) Api.Compiled.Object.EventEdge
edgeSelectionSet =
    SelectionSet.succeed Edge.fromSelectionSet
        |> with EventEdge.cursor
        |> with (SelectionSet.nonNullOrFail <| EventEdge.node eventSelectionSet)


eventSelectionSet : SelectionSet Event Api.Compiled.Object.Event
eventSelectionSet =
    SelectionSet.succeed toEvent
        --        |> with idSelectionSet
        |> with Event.type_
        |> with (Event.user Username.selectionSet)
        |> with (Event.knownHost KnownHost.idSelectionSet)
        |> with (Event.project ProjectId.selectionSet)



--idSelectionSet : SelectionSet Id Api.Compiled.Object.Event
--idSelectionSet =
--    SelectionSet.map Id Event.id


toEvent : String -> Username -> Maybe KnownHost.Id -> Maybe ProjectId.Id -> Event
toEvent type_ username maybeKnownHostId maybeProjectid =
    let
        eventType =
            case type_ of
                "known_host_added" ->
                    Maybe.map KnownHostAdded maybeKnownHostId

                "known_host_verified" ->
                    Maybe.map KnownHostVerified maybeKnownHostId

                "project_added" ->
                    Maybe.map ProjectAdded maybeProjectid

                _ ->
                    Nothing
    in
    eventType
        |> Maybe.withDefault Unknown
        |> Event username



-- VIEW
