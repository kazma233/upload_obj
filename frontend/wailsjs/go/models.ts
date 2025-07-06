export namespace watermark {
	
	export class WatermarkHandle {
	    text: string;
	    size: number;
	    dpi: number;
	    color: string;
	    x: number;
	    y: number;
	    position: string;
	    angle: number;
	
	    static createFrom(source: any = {}) {
	        return new WatermarkHandle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.size = source["size"];
	        this.dpi = source["dpi"];
	        this.color = source["color"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.position = source["position"];
	        this.angle = source["angle"];
	    }
	}

}

