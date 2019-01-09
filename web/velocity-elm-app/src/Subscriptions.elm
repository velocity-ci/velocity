port module Subscriptions exposing (Subscriptions, init, subscriptions, subscribeToKnownHostAdded)

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


subscriptions : (Subscriptions msg -> msg) -> Subscriptions msg -> Sub msg
subscriptions toMsg state =
    Sub.batch
        [ socketStatusConnected (\id -> toMsg <| newSubscriptionStatus ( id, Connected ) state)
        , socketStatusReconnecting (\id -> toMsg <| newSubscriptionStatus ( id, Reconnecting ) state)
        , gotSubscriptionData (\{ id, value } -> newSubscriptionData toMsg { id = id, value = value } state)
        ]


newSubscriptionStatus : ( Int, Status ) -> Subscriptions msg -> Subscriptions msg
newSubscriptionStatus ( id, status ) (Subscriptions subs) =
    subs
        |> Dict.update id (Maybe.map (updateStatus status))
        |> Subscriptions


newSubscriptionData : (Subscriptions msg -> msg) -> { id : Int, value : Value } -> Subscriptions msg -> msg
newSubscriptionData toMsg { id, value } (Subscriptions subs) =
    case Dict.get id subs of
        Just (Subscription _ (KnownHostSubscription selectionSet handler)) ->
            case Decode.decodeValue (Graphql.Document.decoder selectionSet) value of
                Ok data ->
                    handler data

                Err _ ->
                    toMsg (Subscriptions subs)

        Nothing ->
            toMsg (Subscriptions subs)



-- TYPES


type Subscriptions msg
    = Subscriptions (Dict Int (Subscription msg))


type Subscription msg
    = Subscription Status (SubscriptionType msg)


type SubscriptionType msg
    = KnownHostSubscription (SelectionSet KnownHost RootSubscription) (KnownHost -> msg)


type Status
    = NotConnected
    | Connected
    | Reconnecting


init : Subscriptions msg
init =
    Subscriptions Dict.empty


updateStatus : Status -> Subscription msg -> Subscription msg
updateStatus status (Subscription _ type_) =
    Subscription status type_



-- Subscribe to


subscribeToKnownHostAdded : (KnownHost -> msg) -> Subscriptions msg -> ( Subscriptions msg, Cmd msg )
subscribeToKnownHostAdded toMsg state =
    Api.Compiled.Subscription.knownHostAdded KnownHost.selectionSet
        |> knownHostSubscription toMsg state


knownHostSubscription :
    (KnownHost -> msg)
    -> Subscriptions msg
    -> SelectionSet KnownHost RootSubscription
    -> ( Subscriptions msg, Cmd msg )
knownHostSubscription toMsg (Subscriptions internals) selectionSet =
    let
        sub =
            Subscription NotConnected <|
                KnownHostSubscription selectionSet toMsg

        nextId =
            Dict.size internals
    in
        ( Subscriptions (Dict.insert nextId sub internals)
        , subscribeTo ( nextId, Document.serializeSubscription selectionSet )
        )
