export namespace app {
	
	export class BrandInfo {
	    name: string;
	    logo: string;
	
	    static createFrom(source: any = {}) {
	        return new BrandInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.logo = source["logo"];
	    }
	}
	export class PageInfo {
	    title: string;
	    index: number;
	
	    static createFrom(source: any = {}) {
	        return new PageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.index = source["index"];
	    }
	}

}

