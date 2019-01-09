port module Api.Subscriptions exposing (State, init, subscriptions, subscribeToKnownHostAdded)

import Api.Compiled.Subscription
import Dict exposing (Dict)
import Graphql.Document as Document
import Graphql.Operation exposing (RootSubscription)
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)
import KnownHost exposing (KnownHost)
import Json.Encode as Encode exposing (Value)
import Json.Decode as Decode
import Graphql.Document


-- INWARDS PORTs


port gotSubscriptionData : ({ id : Int, value : Value } -> msg) -> Sub msg


port socketStatusConnected : (Int -> msg) -> Sub msg


port socketStatusReconnecting : (Int -> msg) -> Sub msg



-- OUTWARDS PORT


port subscribeTo : ( Int, String ) -> Cmd msg



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
    case Dict.get id subs of
        Just (Subscription _ (KnownHostSubscription selectionSet handler)) ->
            case Decode.decodeValue (Graphql.Document.decoder selectionSet) value of
                Ok data ->
                    handler data

                Err _ ->
                    toMsg (State subs)

        Nothing ->
            toMsg (State subs)



-- TYPES


type State msg
    = State (Dict Int (Subscription msg))


type Subscription msg
    = Subscription Status (SubscriptionType msg)


type SubscriptionType msg
    = KnownHostSubscription (SelectionSet KnownHost RootSubscription) (KnownHost -> msg)


type Status
    = NotConnected
    | Connected
    | Reconnecting


init : State msg
init =
    State Dict.empty


updateStatus : Status -> Subscription msg -> Subscription msg
updateStatus status (Subscription _ type_) =
    Subscription status type_



-- Subscribe to


subscribeToKnownHostAdded : (KnownHost -> msg) -> State msg -> ( State msg, Cmd msg )
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
                KnownHostSubscription selectionSet toMsg

        nextId =
            Dict.size internals
    in
        ( State (Dict.insert nextId sub internals)
        , subscribeTo ( nextId, Document.serializeSubscription selectionSet )
        )
