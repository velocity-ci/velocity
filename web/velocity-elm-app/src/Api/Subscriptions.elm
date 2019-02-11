port module Api.Subscriptions exposing
    ( State
    , StatusMsg
    , init
    , subscribeToEventAdded
    , subscribeToKnownHostAdded
    , subscribeToProjectAdded
    , subscriptions
    )

import Api.Compiled.Subscription
import Dict exposing (Dict)
import Event exposing (Event)
import Graphql.Document as Document
import Graphql.Operation exposing (RootSubscription)
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)
import Json.Decode as Decode
import Json.Encode as Encode exposing (Value)
import KnownHost exposing (KnownHost)
import Project exposing (Project)



-- INWARDS PORTs


port gotSubscriptionData : ({ id : Int, value : Value } -> msg) -> Sub msg


port socketStatusConnected : (Int -> msg) -> Sub msg


port socketStatusReconnecting : (Int -> msg) -> Sub msg



-- OUTWARDS PORT


port subscribeTo : ( Int, String ) -> Cmd msg



-- TYPES


type State msg
    = State (Dict Int (Subscription msg))


type Subscription msg
    = Subscription Status (SubscriptionType msg)


type SubscriptionType msg
    = KnownHostSubscription (SubConfig KnownHost msg)
    | ProjectSubscription (SubConfig Project msg)
    | EventSubscription (SubConfig Event msg)


type alias SubConfig type_ msg =
    { selectionSet : SelectionSet type_ RootSubscription
    , handler : type_ -> msg
    }


type Status
    = NotConnected
    | Connected
    | Reconnecting


type StatusMsg
    = StatusMsg Int Status


init : State msg
init =
    State Dict.empty



-- Subscriptions


subscriptions : (State msg -> msg) -> State msg -> Sub msg
subscriptions toMsg state =
    Sub.batch
        [ socketStatusConnected (\id -> toMsg <| newSubscriptionStatus ( id, Connected ) state)
        , socketStatusReconnecting (\id -> toMsg <| newSubscriptionStatus ( id, Reconnecting ) state)
        , gotSubscriptionData (\{ id, value } -> newSubscriptionData toMsg { id = id, value = value } state)
        ]


newSubscriptionStatus : ( Int, Status ) -> State msg -> State msg
newSubscriptionStatus ( id, status ) (State subs) =
    subs
        |> Dict.update id (Maybe.map (updateStatus status))
        |> State


newSubscriptionData : (State msg -> msg) -> { id : Int, value : Value } -> State msg -> msg
newSubscriptionData toMsg { id, value } (State subs) =
    let
        handleSubscriptionData { selectionSet, handler } =
            case Decode.decodeValue (Document.decoder selectionSet) value of
                Ok data ->
                    handler data

                Err _ ->
                    toMsg (State subs)
    in
    case Dict.get id subs of
        Just (Subscription _ (KnownHostSubscription sub)) ->
            handleSubscriptionData sub

        Just (Subscription _ (ProjectSubscription sub)) ->
            handleSubscriptionData sub

        Just (Subscription _ (EventSubscription sub)) ->
            handleSubscriptionData sub

        Nothing ->
            toMsg (State subs)


updateStatus : Status -> Subscription msg -> Subscription msg
updateStatus status (Subscription _ type_) =
    Subscription status type_



-- Subscribe to


subscribeToEventAdded : (Event -> msg) -> State msg -> ( State msg, Cmd msg )
subscribeToEventAdded toMsg state =
    Api.Compiled.Subscription.eventAdded Event.eventSelectionSet
        |> eventSubscription toMsg state


eventSubscription :
    (Event -> msg)
    -> State msg
    -> SelectionSet Event RootSubscription
    -> ( State msg, Cmd msg )
eventSubscription toMsg (State internals) selectionSet =
    let
        sub =
            Subscription NotConnected <|
                EventSubscription { selectionSet = selectionSet, handler = toMsg }

        nextId =
            Dict.size internals
    in
    ( State (Dict.insert nextId sub internals)
    , subscribeTo ( nextId, Document.serializeSubscription selectionSet )
    )


subscribeToProjectAdded : (Project -> msg) -> State msg -> ( State msg, Cmd msg )
subscribeToProjectAdded toMsg state =
    Api.Compiled.Subscription.projectAdded Project.selectionSet
        |> projectSubscription toMsg state


projectSubscription :
    (Project -> msg)
    -> State msg
    -> SelectionSet Project RootSubscription
    -> ( State msg, Cmd msg )
projectSubscription toMsg (State internals) selectionSet =
    let
        sub =
            Subscription NotConnected <|
                ProjectSubscription { selectionSet = selectionSet, handler = toMsg }

        nextId =
            Dict.size internals
    in
    ( State (Dict.insert nextId sub internals)
    , subscribeTo ( nextId, Document.serializeSubscription selectionSet )
    )


subscribeToKnownHostAdded : (KnownHost -> msg2) -> State msg2 -> ( State msg2, Cmd msg2 )
subscribeToKnownHostAdded toMsg state =
    Api.Compiled.Subscription.knownHostAdded KnownHost.selectionSet
        |> knownHostSubscription toMsg state


knownHostSubscription :
    (KnownHost -> msg)
    -> State msg
    -> SelectionSet KnownHost RootSubscription
    -> ( State msg, Cmd msg )
knownHostSubscription toMsg (State internals) selectionSet =
    let
        sub =
            Subscription NotConnected <|
                KnownHostSubscription { selectionSet = selectionSet, handler = toMsg }

        nextId =
            Dict.size internals
    in
    ( State (Dict.insert nextId sub internals)
    , subscribeTo ( nextId, Document.serializeSubscription selectionSet )
    )
