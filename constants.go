package wsman

/*
Copyright 2015 Victor Lowther <victor.lowther@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

const (
	// Models any simple single item retrieval
	GET = "http://schemas.xmlsoap.org/ws/2004/09/transfer/Get"

	// Models an update of an entire item
	PUT = "http://schemas.xmlsoap.org/ws/2004/09/transfer/Put"

	// Models creation of a new item
	CREATE = "http://schemas.xmlsoap.org/ws/2004/09/transfer/Create"

	// Models the deletion of an item
	DELETE = "http://schemas.xmlsoap.org/ws/2004/09/transfer/Delete"

	// Begins an enumeration or query
	ENUMERATE = "http://schemas.xmlsoap.org/ws/2004/09/enumeration/Enumerate"

	// Retrieves the next batch of results from enumeration
	PULL = "http://schemas.xmlsoap.org/ws/2004/09/enumeration/Pull"

	// Releases an active enumerator
	RELEASE = "http://schemas.xmlsoap.org/ws/2004/09/enumeration/Release"

	// Models a subscription to an event source
	SUBSCRIBE = "http://schemas.xmlsoap.org/ws/2004/08/eventing/Subscribe"

	// Renews a subscription prior to its expiration
	RENEW = "http://schemas.xmlsoap.org/ws/2004/08/eventing/Renew"

	// Requests the status of a subscription
	GET_STATUS = "http://schemas.xmlsoap.org/ws/2004/08/eventing/GetStatus"

	// Removes an active subscription
	UNSUBSCRIBE = "http://schemas.xmlsoap.org/ws/2004/08/eventing/Unsubscribe"

	// Delivers a message to indicate that a subscription has terminated
	SUBSCRIBE_END = "http://schemas.xmlsoap.org/ws/2004/08/eventing/SubscriptionEnd"

	// Delivers batched events based on a subscription
	EVENTS = "http://schemas.dmtf.org/wbem/wsman/1/wsman/Events"

	// A pseudo-event that models a heartbeat of an active subscription;
	// delivered when no real events are available, but used to indicate that the
	// event subscription and delivery mechanism is still active
	HEARTBEAT = "http://schemas.dmtf.org/wbem/wsman/1/wsman/Heartbeat"

	// A pseudo-event that indicates that the real event was dropped
	DROPPED_EVENTS = "http://schemas.dmtf.org/wbem/wsman/1/wsman/DroppedEvents"

	// Used by event subscribers to acknowledge receipt of events;
	// allows event streams to be strictly sequenced
	ACK = "http://schemas.dmtf.org/wbem/wsman/1/wsman/Ack"

	// Used for a singleton event that does not define its own action
	EVENT = "http://schemas.dmtf.org/wbem/wsman/1/wsman/Event"
)
