module Event exposing (Log, selectionSet, view)

import Api.Compiled.Object
import Api.Compiled.Object.Event as Event
import Api.Compiled.Object.EventConnection as EventConnection
import Api.Compiled.Object.EventEdge as EventEdge
import Connection exposing (Connection)
import Edge exposing (Edge)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Font as Font
import GitUrl
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)
import Icon
import KnownHost exposing (KnownHost)
import PageInfo exposing (PageInfo)
import Palette
import Project exposing (Project)
import Project.Id as ProjectId
import Set exposing (Set)
import Time
import Username exposing (Username)


type Log
    = Log Internals


type alias Internals =
    { seen : Set Id
    , events : Connection Event
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
toEvent type_ username maybeKnownHostId maybeProjectId =
    let
        eventType =
            case type_ of
                "known_host_created" ->
                    Maybe.map KnownHostAdded maybeKnownHostId

                "known_host_verified" ->
                    Maybe.map KnownHostVerified maybeKnownHostId

                "project_created" ->
                    Maybe.map ProjectAdded maybeProjectId

                _ ->
                    Nothing
    in
    eventType
        |> Maybe.withDefault Unknown
        |> Event username



-- VIEW


type alias ViewConfiguration =
    { projects : List Project
    , knownHosts : List KnownHost
    , maybeLog : Maybe Log
    }


view : ViewConfiguration -> Element msg
view config =
    case config.maybeLog of
        Just log ->
            viewLogContainer log config

        Nothing ->
            none


viewLogContainer : Log -> ViewConfiguration -> Element msg
viewLogContainer (Log internals) { projects, knownHosts } =
    let
        events =
            internals.events.edges
                |> List.map (Edge.node >> viewItem)

        viewItem =
            viewLogItem projects knownHosts
    in
    column
        [ Background.color Palette.primary2
        , width fill
        , height fill
        , paddingEach { top = 80, bottom = 90, left = 20, right = 20 }
        , spacing 10
        ]
        (viewNotificationsPanelHeading :: events)


viewNotificationsPanelHeading : Element msg
viewNotificationsPanelHeading =
    row
        [ width fill
        , Font.color Palette.neutral7
        , Font.extraLight
        , Font.size 17
        ]
        [ text "Recent activity" ]


viewLogItem : List Project -> List KnownHost -> Event -> Element msg
viewLogItem projects knownHosts (Event username type_) =
    case type_ of
        KnownHostAdded id ->
            viewKnownHostAdded knownHosts username id

        KnownHostVerified id ->
            viewKnownHostVerified knownHosts username id

        ProjectAdded id ->
            viewProjectAdded projects username id

        Unknown ->
            none


viewKnownHostAdded : List KnownHost -> Username -> KnownHost.Id -> Element msg
viewKnownHostAdded knownHosts username id =
    case KnownHost.find knownHosts id of
        Just knownHost ->
            let
                icon =
                    KnownHost.host knownHost
                        |> GitUrl.sourceIcon
            in
            viewLogItemContainer
                [ el [ width (px 40), height (px 40) ] <|
                    icon Icon.fullSizeOptions
                , el [ width fill ]
                    (paragraph
                        [ Font.size 15
                        , alignLeft
                        , paddingXY 10 0
                        ]
                        [ el [ alignLeft, Font.color Palette.neutral5 ] (text "Known host for ")
                        , el [ Font.heavy, alignLeft, Font.color Palette.neutral6 ] (text <| KnownHost.host knownHost)
                        , el [ alignLeft, Font.color Palette.neutral5 ] (text " created by ")
                        , el [ Font.heavy, alignLeft, Font.color Palette.neutral6 ] (text <| Username.toString username)
                        , el [ Font.extraLight, Font.size 13, Font.color Palette.neutral5, alignLeft ] (text " 5 mins ago")
                        ]
                    )
                ]

        Nothing ->
            none


viewKnownHostVerified : List KnownHost -> Username -> KnownHost.Id -> Element msg
viewKnownHostVerified knownHosts username id =
    case KnownHost.find knownHosts id of
        Just knownHost ->
            let
                icon =
                    Icon.checkCircle
            in
            viewLogItemContainer
                [ el [ width (px 40), height (px 40) ] <|
                    icon Icon.fullSizeOptions
                , el [ width fill ]
                    (paragraph
                        [ Font.size 15
                        , alignLeft
                        , paddingXY 10 0
                        ]
                        [ el [ alignLeft, Font.color Palette.neutral5 ] (text "Known host for ")
                        , el [ Font.heavy, alignLeft, Font.color Palette.neutral6 ] (text <| KnownHost.host knownHost)
                        , el [ alignLeft, Font.color Palette.neutral5 ] (text " verified by ")
                        , el [ Font.heavy, alignLeft, Font.color Palette.neutral6 ] (text <| Username.toString username)
                        , el [ Font.extraLight, Font.size 13, Font.color Palette.neutral5, alignLeft ] (text " 5 mins ago")
                        ]
                    )
                ]

        Nothing ->
            none


viewProjectAdded : List Project -> Username -> ProjectId.Id -> Element msg
viewProjectAdded projects username id =
    case Project.find projects id of
        Just project ->
            viewLogItemContainer
                [ el [ width (px 40), height (px 40) ] <| Project.thumbnail project
                , el [ width fill ]
                    (paragraph
                        [ Font.size 15
                        , alignLeft
                        , paddingXY 10 0
                        ]
                        [ el [ alignLeft, Font.color Palette.neutral5 ] (text "Project ")
                        , el [ Font.heavy, alignLeft, Font.color Palette.neutral6 ] (text <| Project.name project)
                        , el [ alignLeft, Font.color Palette.neutral5 ] (text " created ")
                        , el [ Font.heavy, alignLeft, Font.color Palette.neutral6 ] (text <| Username.toString username)
                        , el [ Font.extraLight, Font.size 13, Font.color Palette.neutral5, alignLeft ] (text " 8 hours ago")
                        ]
                    )
                ]

        Nothing ->
            none


viewLogItemContainer : List (Element msg) -> Element msg
viewLogItemContainer children =
    row
        [ Border.width 1
        , Border.color Palette.primary4
        , Font.color Palette.neutral4
        , Font.light
        , Border.dashed
        , Border.rounded 5
        , padding 10
        , width fill
        , mouseOver [ Background.color Palette.primary3, Font.color Palette.neutral5 ]
        ]
        children



--viewProjectNotification : List Project -> CategoryProject -> Element msg
--viewProjectNotification projects category =
--    let
--        maybeProject =
--            projectFromCategory projects category
--    in
--        case maybeProject of
--            Just project ->
--                row
--                    [ Border.width 1
--                    , Border.color Palette.primary4
--                    , Font.color Palette.neutral4
--                    , Font.light
--                    , Border.dashed
--                    , Border.rounded 5
--                    , padding 10
--                    , width fill
--                    , mouseOver [ Background.color Palette.primary3, Font.color Palette.neutral5 ]
--                    ]
--                    [ el [ width (px 40), height (px 40) ] <| Project.thumbnail project
--                    , el [ width fill ]
--                        (paragraph
--                            [ Font.size 15
--                            , alignLeft
--                            , paddingXY 10 0
--                            ]
--                            [ el [ alignLeft, Font.color Palette.neutral5 ] (text "Project ")
--                            , el [ Font.heavy, alignLeft, Font.color Palette.neutral6 ] (text <| Project.name project)
--                            , el [ alignLeft, Font.color Palette.neutral5 ] (text " created ")
--                            , el [ Font.extraLight, Font.size 13, Font.color Palette.neutral5, alignLeft ] (text "8 hours ago")
--                            ]
--                        )
--                    ]
--
--            Nothing ->
--                none
--
