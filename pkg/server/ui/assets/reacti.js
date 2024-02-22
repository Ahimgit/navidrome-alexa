function findComponentWithProp(id, propName) {
    function traverseFiber(fiber, propName) {
        if (!fiber) {
            return null;
        }
        if (fiber.memoizedProps && fiber.memoizedProps[propName] !== undefined) {
            return fiber;
        }
        let foundComponent = traverseFiber(fiber.child, propName);
        if (foundComponent) {
            return foundComponent;
        }
        return traverseFiber(fiber.sibling, propName);
    }
    const rootFiber = document.getElementById(id)._reactRootContainer._internalRoot.current;
    return traverseFiber(rootFiber, propName);
}

const component = findComponentWithProp('root', 'showLyric');


const aL_Orig = component.stateNode.audio.load;
component.stateNode.audio.load = function() {
	console.log('load', this.src);
	//return aL_Orig.call(component.stateNode.audio);
	return Promise.resolve();
}


component.stateNode.audio.pause = function() { 
	console.log('pause', this.src);
	this.__paused = true;
	//component.stateNode.onAudioPause();
}

component.stateNode.audio.play = function() {
	//component.stateNode.onTogglePlay();
	//setTimeout(() => {
      //this.currentTime=this.currentTime+0.5;
    //}, 500)
    //component.stateNode.onAudioPlay();
	console.log('play');
	this.__paused = false;
	//return aP_Orig.call(component.stateNode.audio);
	if(!this.__ticks) {
		this.__ticks = true;
		setInterval(() => {
			if(!this.__paused) {
                    console.log("tick")
					component.memoizedState.currentTime += 1;
                    if(component.stateNode.audio.duration <= component.memoizedState.currentTime) {
                        component.stateNode.audio.currentTime = component.stateNode.audio.duration;
                    }
                    component.stateNode.forceUpdate();
			}
		}, 1000)
	}
	return Promise.resolve();
}

Object.defineProperty(component.stateNode.audio, 'paused', {
    get() {
        console.log('paused get', this.__paused);
        return this.__paused;//originalGetter.call(component.stateNode.audio);
    },
    set(newValue) { },
    configurable: true // This needs to be true to redefine a property
});






const originalComponentDidUpdate = component.stateNode.componentDidUpdate;
component.stateNode.componentDidUpdate = (...args) => {
    if (originalComponentDidUpdate) {
        originalComponentDidUpdate.apply(component.stateNode, args);
    }
    // Reapply the mock
    component.stateNode.audio = mockAudioElement;
};

// update time on mock
component.stateNode.audio.currentTime = 40; component.stateNode.audioTimeUpdate()




component.stateNode.props.playIndex=0;
component.memoizedProps = { ...component.memoizedProps }
component.stateNode.forceUpdate();

component.memoizedProps = { ...component.memoizedProps }
component.stateNode.forceUpdate();


component.memoizedState.currentTime=100;
component.memoizedState = { ... component.memoizedState }
component.stateNode.forceUpdate();


component.stateNode.playByIndex(5)
component.stateNode.props.audioLists[0].trackId

component.stateNode.audio.currentTime = 6.004;
component.stateNode.setAudioVolume(0)
component.stateNode.audio.pause();
component.stateNode.audio.play();

let zov = component.stateNode.onPlayNextAudio; component.stateNode.onPlayNextAudio = ()=>{ console.log(123); zov(); }




component.stateNode.forceUpdate();

