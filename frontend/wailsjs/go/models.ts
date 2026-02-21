export namespace main {
	
	export class CTEventService {
	    id: number;
	    serviceId: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new CTEventService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.serviceId = source["serviceId"];
	        this.name = source["name"];
	    }
	}
	export class CTEvent {
	    id: number;
	    name: string;
	    startDate: string;
	    eventServices: CTEventService[];
	
	    static createFrom(source: any = {}) {
	        return new CTEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.startDate = source["startDate"];
	        this.eventServices = this.convertValues(source["eventServices"], CTEventService);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class CTFact {
	    id: number;
	    title: string;
	
	    static createFrom(source: any = {}) {
	        return new CTFact(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	    }
	}
	export class CTFactDefinition {
	    id: number;
	    name: string;
	    nameTranslated: string;
	    sortKey: number;
	    type: string;
	    options: string[];
	
	    static createFrom(source: any = {}) {
	        return new CTFactDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.nameTranslated = source["nameTranslated"];
	        this.sortKey = source["sortKey"];
	        this.type = source["type"];
	        this.options = source["options"];
	    }
	}
	export class CTService {
	    id: number;
	    name: string;
	    serviceGroupId: number;
	    sortKey: number;
	
	    static createFrom(source: any = {}) {
	        return new CTService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.serviceGroupId = source["serviceGroupId"];
	        this.sortKey = source["sortKey"];
	    }
	}
	export class CTServiceGroup {
	    id: number;
	    name: string;
	    sortKey: number;
	
	    static createFrom(source: any = {}) {
	        return new CTServiceGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sortKey = source["sortKey"];
	    }
	}
	export class CTMasterData {
	    serviceGroups: CTServiceGroup[];
	    services: CTService[];
	    facts: CTFact[];
	
	    static createFrom(source: any = {}) {
	        return new CTMasterData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceGroups = this.convertValues(source["serviceGroups"], CTServiceGroup);
	        this.services = this.convertValues(source["services"], CTService);
	        this.facts = this.convertValues(source["facts"], CTFact);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class Dienst {
	    id: string;
	    label: string;
	    persons: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Dienst(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.persons = source["persons"];
	    }
	}
	export class Einsatz {
	    datum: string;
	    zeit?: string;
	    musikteam: string;
	    event_id?: number;
	    event_name?: string;
	
	    static createFrom(source: any = {}) {
	        return new Einsatz(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.datum = source["datum"];
	        this.zeit = source["zeit"];
	        this.musikteam = source["musikteam"];
	        this.event_id = source["event_id"];
	        this.event_name = source["event_name"];
	    }
	}
	export class EventFactDisplay {
	    factId: number;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new EventFactDisplay(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.factId = source["factId"];
	        this.value = source["value"];
	    }
	}
	export class Person {
	    id: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Person(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}

}

